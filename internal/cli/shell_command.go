package cli

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
)

const shellUsage = "usage: goscrapling shell -c <script> [--loglevel <level>]"

var (
	shellGetCall     = regexp.MustCompile(`^get\((.*)\)$`)
	shellPrintCall   = regexp.MustCompile(`^print\((.*)\)$`)
	shellLenCSSCall  = regexp.MustCompile(`^len\((page|response)\.css\((.*)\)\)$`)
	shellCSSGetCall  = regexp.MustCompile(`^(page|response)\.css\((.*)\)\.get\((.*)\)$`)
	shellCSSTextCall = regexp.MustCompile(`^(page|response)\.css\((.*)\)\.text\(\)$`)
)

type shellPlan struct {
	code string
}

type shellSession struct {
	fetcher goscrapling.Fetcher
	page    *goscrapling.Response
	pages   []*goscrapling.Response
}

func runShell(stdout io.Writer, args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, err := fmt.Fprintln(stdout, shellUsage)
		return err
	}
	plan, err := parseShellPlan(args)
	if err != nil {
		return err
	}
	return (&shellSession{}).run(stdout, plan.code)
}

func parseShellPlan(args []string) (shellPlan, error) {
	var plan shellPlan
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "--code":
			value, ok := nextArg(args, &i, args[i])
			if !ok {
				return shellPlan{}, parseError("%s requires a value", args[i])
			}
			plan.code = value
		case "-L", "--loglevel":
			if _, ok := nextArg(args, &i, args[i]); !ok {
				return shellPlan{}, parseError("%s requires a value", args[i])
			}
		default:
			return shellPlan{}, parseError("unknown shell option %q", args[i])
		}
	}
	if strings.TrimSpace(plan.code) == "" {
		return shellPlan{}, parseError("shell requires -c <script>; interactive REPL is not implemented")
	}
	return plan, nil
}

func (s *shellSession) run(stdout io.Writer, script string) error {
	for _, statement := range splitShellStatements(script) {
		if statement == "" {
			continue
		}
		if match := shellGetCall.FindStringSubmatch(statement); match != nil {
			url, err := parseShellStringArg(match[1])
			if err != nil {
				return parseError("get requires a quoted URL")
			}
			if _, err := s.get(url); err != nil {
				return err
			}
			continue
		}
		if match := shellPrintCall.FindStringSubmatch(statement); match != nil {
			value, err := s.eval(match[1])
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(stdout, value); err != nil {
				return err
			}
			continue
		}
		return parseError("unsupported shell statement %q", statement)
	}
	return nil
}

func (s *shellSession) get(rawURL string) (*goscrapling.Response, error) {
	response, err := s.fetcher.Get(rawURL, goscrapling.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("shell get %q: %w", rawURL, err)
	}
	s.page = response
	s.pages = append(s.pages, response)
	if len(s.pages) > 5 {
		s.pages = append([]*goscrapling.Response(nil), s.pages[len(s.pages)-5:]...)
	}
	return response, nil
}

func (s *shellSession) eval(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "page.url", "response.url":
		page, err := s.currentPage()
		if err != nil {
			return "", err
		}
		return page.URL(), nil
	case "page.status", "response.status":
		page, err := s.currentPage()
		if err != nil {
			return "", err
		}
		return strconv.Itoa(page.StatusCode()), nil
	case "len(pages)":
		return strconv.Itoa(len(s.pages)), nil
	}

	if match := shellLenCSSCall.FindStringSubmatch(expr); match != nil {
		page, err := s.currentPage()
		if err != nil {
			return "", err
		}
		selector, err := parseShellStringArg(match[2])
		if err != nil {
			return "", parseError("css requires a quoted selector")
		}
		return strconv.Itoa(page.CSS(selector).Len()), nil
	}
	if match := shellCSSGetCall.FindStringSubmatch(expr); match != nil {
		page, err := s.currentPage()
		if err != nil {
			return "", err
		}
		selector, err := parseShellStringArg(match[2])
		if err != nil {
			return "", parseError("css requires a quoted selector")
		}
		defaultValue := ""
		if trimmed := strings.TrimSpace(match[3]); trimmed != "" {
			parsed, err := parseShellStringArg(trimmed)
			if err != nil {
				return "", parseError("get default requires a quoted string")
			}
			defaultValue = parsed
		}
		return page.CSS(selector).Get(defaultValue).String(), nil
	}
	if match := shellCSSTextCall.FindStringSubmatch(expr); match != nil {
		page, err := s.currentPage()
		if err != nil {
			return "", err
		}
		selector, err := parseShellStringArg(match[2])
		if err != nil {
			return "", parseError("css requires a quoted selector")
		}
		return page.CSS(selector).Text(), nil
	}
	return "", parseError("unsupported shell expression %q", expr)
}

func (s *shellSession) currentPage() (*goscrapling.Response, error) {
	if s.page == nil {
		return nil, parseError("page is not set; call get(url) first")
	}
	return s.page, nil
}

func parseShellStringArg(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 2 {
		return "", fmt.Errorf("missing quotes")
	}
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
		return strconv.Unquote(raw)
	}
	if strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'") {
		inner := strings.TrimSuffix(strings.TrimPrefix(raw, "'"), "'")
		inner = strings.ReplaceAll(inner, `\\`, `\`)
		inner = strings.ReplaceAll(inner, `\'`, `'`)
		return inner, nil
	}
	return "", fmt.Errorf("unquoted string")
}

func splitShellStatements(script string) []string {
	var statements []string
	var current strings.Builder
	var quote rune
	escaped := false
	for _, r := range script {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && quote != 0 {
			current.WriteRune(r)
			escaped = true
			continue
		}
		if quote != 0 {
			current.WriteRune(r)
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
			current.WriteRune(r)
		case ';', '\n':
			statements = append(statements, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	statements = append(statements, strings.TrimSpace(current.String()))
	return statements
}
