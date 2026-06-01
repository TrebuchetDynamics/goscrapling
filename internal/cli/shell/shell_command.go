package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/arguments"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/shell/curlcommand"
)

const shellUsage = "usage: goscrapling shell -c <script> [--loglevel <level>]"

var (
	shellStaticMethodCall = regexp.MustCompile(`^(get|post|put|delete)\((.*)\)$`)
	shellPrintCall        = regexp.MustCompile(`^print\((.*)\)$`)
	shellLenCSSCall       = regexp.MustCompile(`^len\((page|response)\.css\((.*)\)\)$`)
	shellCSSGetCall       = regexp.MustCompile(`^(page|response)\.css\((.*)\)\.get\((.*)\)$`)
	shellCSSTextCall      = regexp.MustCompile(`^(page|response)\.css\((.*)\)\.text\(\)$`)
	shellCurl2FetcherCall = regexp.MustCompile(`^curl2fetcher\((.*)\)$`)
	shellHelpCall         = regexp.MustCompile(`^help\(\)$`)
	shellViewCall         = regexp.MustCompile(`^view\((page|response)\)$`)
	shellUncurlFieldCall  = regexp.MustCompile(`^uncurl\((.*)\)\.(method|url|body)$`)
	shellPagesFieldCall   = regexp.MustCompile(`^pages\[(-?\d+)\]\.(url|status)$`)
	shellUncurlValueCall  = regexp.MustCompile(`^uncurl\((.*)\)\.(header|cookie|param)\((.*)\)$`)
)

type shellPlan struct {
	code string
}

type shellSession struct {
	fetcher goscrapling.Fetcher
	page    *goscrapling.Response
	pages   []*goscrapling.Response
}

type shellCallOptions struct {
	headers http.Header
	body    string
}

func Run(stdout io.Writer, args []string) error {
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
		if match := shellStaticMethodCall.FindStringSubmatch(statement); match != nil {
			url, opts, err := parseShellMethodArgs(match[1], match[2])
			if err != nil {
				return err
			}
			if _, err := s.staticMethod(match[1], url, opts); err != nil {
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
		if match := shellCurl2FetcherCall.FindStringSubmatch(statement); match != nil {
			curlCommand, err := parseShellStringArg(match[1])
			if err != nil {
				return parseError("curl2fetcher requires a quoted curl command")
			}
			if _, err := s.curl2fetcher(curlCommand); err != nil {
				return err
			}
			continue
		}
		if shellHelpCall.MatchString(statement) {
			if _, err := fmt.Fprint(stdout, shellHelpText()); err != nil {
				return err
			}
			continue
		}
		if shellViewCall.MatchString(statement) {
			path, err := s.writeViewArtifact()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(stdout, "view wrote %s\n", path); err != nil {
				return err
			}
			continue
		}
		return parseError("unsupported shell statement %q", statement)
	}
	return nil
}

func shellHelpText() string {
	return `-> Available goscrapling shell objects:
   - Fetcher-style static request shortcuts
   - Response selector helpers on page and response

-> Useful shortcuts:
   - get                            Shortcut for static GET
   - post                           Shortcut for static POST
   - put                            Shortcut for static PUT
   - delete                         Shortcut for static DELETE

-> Useful commands:
   - page / response                The response object of the last page fetched
   - pages                          The last 5 response objects fetched
   - uncurl('curl_command')         Convert a DevTools-style curl command to request metadata
   - curl2fetcher('curl_command')   Convert a curl command and execute it with the static fetcher
   - help()                         Show this help message

Note: interactive REPL is not implemented in goscrapling; use shell -c with scripted statements.
`
}

func (s *shellSession) staticMethod(method string, rawURL string, callOpts shellCallOptions) (*goscrapling.Response, error) {
	opts := goscrapling.RequestOptions{Headers: callOpts.headers}
	if callOpts.body != "" {
		opts.Body = strings.NewReader(callOpts.body)
	}

	var response *goscrapling.Response
	var err error
	switch method {
	case "get":
		response, err = s.fetcher.Get(rawURL, opts)
	case "post":
		response, err = s.fetcher.Post(rawURL, opts)
	case "put":
		response, err = s.fetcher.Put(rawURL, opts)
	case "delete":
		response, err = s.fetcher.Delete(rawURL, opts)
	default:
		return nil, parseError("unsupported shell method %q", method)
	}
	if err != nil {
		return nil, fmt.Errorf("shell %s %q: %w", method, rawURL, err)
	}
	s.updatePage(response)
	return response, nil
}

func (s *shellSession) curl2fetcher(curlCommand string) (*goscrapling.Response, error) {
	request, err := curlcommand.Parse(curlCommand)
	if err != nil {
		return nil, err
	}
	opts := goscrapling.RequestOptions{
		Headers:      request.Headers,
		CookieValues: request.Cookies,
		Params:       request.Params,
	}
	if request.Body != "" {
		opts.Body = strings.NewReader(request.Body)
		if opts.Headers.Get("Content-Type") == "" {
			opts.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	var response *goscrapling.Response
	switch request.Method {
	case "get":
		response, err = s.fetcher.Get(request.URL, opts)
	case "post":
		response, err = s.fetcher.Post(request.URL, opts)
	case "put":
		response, err = s.fetcher.Put(request.URL, opts)
	case "delete":
		response, err = s.fetcher.Delete(request.URL, opts)
	default:
		return nil, parseError("unsupported curl method %q", request.Method)
	}
	if err != nil {
		return nil, fmt.Errorf("shell curl2fetcher %q: %w", request.URL, err)
	}
	s.updatePage(response)
	return response, nil
}

func (s *shellSession) updatePage(response *goscrapling.Response) {
	s.page = response
	s.pages = append(s.pages, response)
	if len(s.pages) > 5 {
		s.pages = append([]*goscrapling.Response(nil), s.pages[len(s.pages)-5:]...)
	}
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

	if match := shellPagesFieldCall.FindStringSubmatch(expr); match != nil {
		page, err := s.pageAt(match[1])
		if err != nil {
			return "", err
		}
		switch match[2] {
		case "url":
			return page.URL(), nil
		case "status":
			return strconv.Itoa(page.StatusCode()), nil
		}
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
	if match := shellUncurlFieldCall.FindStringSubmatch(expr); match != nil {
		request, err := parseShellUncurlArg(match[1])
		if err != nil {
			return "", err
		}
		switch match[2] {
		case "method":
			return request.Method, nil
		case "url":
			return request.URL, nil
		case "body":
			return request.Body, nil
		}
	}
	if match := shellUncurlValueCall.FindStringSubmatch(expr); match != nil {
		request, err := parseShellUncurlArg(match[1])
		if err != nil {
			return "", err
		}
		key, err := parseShellStringArg(match[3])
		if err != nil {
			return "", parseError("uncurl %s requires a quoted key", match[2])
		}
		switch match[2] {
		case "header":
			return request.Headers.Get(key), nil
		case "cookie":
			return request.Cookies[key], nil
		case "param":
			return request.Params.Get(key), nil
		}
	}
	return "", parseError("unsupported shell expression %q", expr)
}

func parseShellUncurlArg(raw string) (curlcommand.Request, error) {
	curlCommand, err := parseShellStringArg(raw)
	if err != nil {
		return curlcommand.Request{}, parseError("uncurl requires a quoted curl command")
	}
	return curlcommand.Parse(curlCommand)
}

func parseShellMethodArgs(method string, raw string) (string, shellCallOptions, error) {
	parts := splitShellArguments(raw)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "", shellCallOptions{}, parseError("%s requires a quoted URL", method)
	}
	url, err := parseShellStringArg(parts[0])
	if err != nil {
		return "", shellCallOptions{}, parseError("%s requires a quoted URL", method)
	}
	opts := shellCallOptions{headers: http.Header{}}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || strings.TrimSpace(key) == "" {
			return "", shellCallOptions{}, parseError("%s option must be key=value", method)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "data":
			body, err := parseShellStringArg(value)
			if err != nil {
				return "", shellCallOptions{}, parseError("data requires a quoted string")
			}
			opts.body = body
			if opts.headers.Get("Content-Type") == "" {
				opts.headers.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		case "json":
			if !json.Valid([]byte(value)) {
				return "", shellCallOptions{}, parseError("json requires a valid JSON object or array")
			}
			opts.body = value
			if opts.headers.Get("Content-Type") == "" {
				opts.headers.Set("Content-Type", "application/json")
			}
		case "headers":
			headers, err := parseShellHeaders(value)
			if err != nil {
				return "", shellCallOptions{}, err
			}
			for headerKey, values := range headers {
				for _, headerValue := range values {
					opts.headers.Add(headerKey, headerValue)
				}
			}
		default:
			return "", shellCallOptions{}, parseError("unsupported %s option %q", method, key)
		}
	}
	return url, opts, nil
}

func parseShellHeaders(raw string) (http.Header, error) {
	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, parseError("headers requires a JSON object of string values")
	}
	headers := http.Header{}
	for key, value := range parsed {
		headers.Add(key, value)
	}
	return headers, nil
}

func splitShellArguments(raw string) []string {
	var parts []string
	var current strings.Builder
	var quote rune
	escaped := false
	braceDepth := 0
	bracketDepth := 0
	for _, r := range raw {
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
		case '{':
			braceDepth++
			current.WriteRune(r)
		case '}':
			if braceDepth > 0 {
				braceDepth--
			}
			current.WriteRune(r)
		case '[':
			bracketDepth++
			current.WriteRune(r)
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			current.WriteRune(r)
		case ',':
			if braceDepth == 0 && bracketDepth == 0 {
				parts = append(parts, strings.TrimSpace(current.String()))
				current.Reset()
				continue
			}
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}
	parts = append(parts, strings.TrimSpace(current.String()))
	return parts
}

func (s *shellSession) currentPage() (*goscrapling.Response, error) {
	if s.page == nil {
		return nil, parseError("page is not set; call get(url) first")
	}
	return s.page, nil
}

func (s *shellSession) writeViewArtifact() (string, error) {
	page, err := s.currentPage()
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "goscrapling_view_*.html")
	if err != nil {
		return "", fmt.Errorf("write view artifact: %w", err)
	}
	path := file.Name()
	if _, err := file.Write(page.Body()); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write view artifact: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("write view artifact: %w", err)
	}
	return path, nil
}

func (s *shellSession) pageAt(rawIndex string) (*goscrapling.Response, error) {
	index, err := strconv.Atoi(rawIndex)
	if err != nil {
		return nil, parseError("pages index must be an integer")
	}
	if index < 0 {
		index = len(s.pages) + index
	}
	if index < 0 || index >= len(s.pages) {
		return nil, parseError("pages index %s out of range", rawIndex)
	}
	return s.pages[index], nil
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

func parseError(format string, args ...any) error {
	return diagnostics.ParseError(format, args...)
}

func nextArg(args []string, index *int, name string) (string, bool) {
	return arguments.NextValue(args, index, name)
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
