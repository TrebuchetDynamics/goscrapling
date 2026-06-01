package shell

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/arguments"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
)

const shellUsage = "usage: goscrapling shell -c <script> [--loglevel <level>]"

var (
	shellGetCall          = regexp.MustCompile(`^get\((.*)\)$`)
	shellPrintCall        = regexp.MustCompile(`^print\((.*)\)$`)
	shellLenCSSCall       = regexp.MustCompile(`^len\((page|response)\.css\((.*)\)\)$`)
	shellCSSGetCall       = regexp.MustCompile(`^(page|response)\.css\((.*)\)\.get\((.*)\)$`)
	shellCSSTextCall      = regexp.MustCompile(`^(page|response)\.css\((.*)\)\.text\(\)$`)
	shellCurl2FetcherCall = regexp.MustCompile(`^curl2fetcher\((.*)\)$`)
	shellUncurlFieldCall  = regexp.MustCompile(`^uncurl\((.*)\)\.(method|url|body)$`)
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

type shellCurlRequest struct {
	Method  string
	URL     string
	Headers http.Header
	Cookies map[string]string
	Params  url.Values
	Body    string
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
		return parseError("unsupported shell statement %q", statement)
	}
	return nil
}

func (s *shellSession) get(rawURL string) (*goscrapling.Response, error) {
	response, err := s.fetcher.Get(rawURL, goscrapling.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("shell get %q: %w", rawURL, err)
	}
	s.updatePage(response)
	return response, nil
}

func (s *shellSession) curl2fetcher(curlCommand string) (*goscrapling.Response, error) {
	request, err := parseShellCurlCommand(curlCommand)
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

func parseShellUncurlArg(raw string) (shellCurlRequest, error) {
	curlCommand, err := parseShellStringArg(raw)
	if err != nil {
		return shellCurlRequest{}, parseError("uncurl requires a quoted curl command")
	}
	return parseShellCurlCommand(curlCommand)
}

func parseShellCurlCommand(command string) (shellCurlRequest, error) {
	tokens, err := splitShellWords(command)
	if err != nil {
		return shellCurlRequest{}, parseError("curl command parse error: %v", err)
	}
	if len(tokens) > 0 && tokens[0] == "curl" {
		tokens = tokens[1:]
	}
	request := shellCurlRequest{
		Method:  "get",
		Headers: http.Header{},
		Cookies: map[string]string{},
		Params:  url.Values{},
	}
	var rawURL string
	var bodyForParams bool
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token {
		case "-X", "--request":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return shellCurlRequest{}, parseError("%s requires a value", token)
			}
			request.Method = strings.ToLower(value)
		case "-H", "--header":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return shellCurlRequest{}, parseError("%s requires a value", token)
			}
			key, headerValue, err := parseHeader(value)
			if err != nil {
				return shellCurlRequest{}, err
			}
			if strings.EqualFold(key, "Cookie") {
				mergeShellCookies(request.Cookies, headerValue)
			} else {
				request.Headers.Add(key, headerValue)
			}
		case "-b", "--cookie":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return shellCurlRequest{}, parseError("%s requires a value", token)
			}
			mergeShellCookies(request.Cookies, value)
		case "-d", "--data", "--data-raw", "--data-binary":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return shellCurlRequest{}, parseError("%s requires a value", token)
			}
			request.Body = strings.TrimPrefix(value, "$")
			if request.Method == "get" {
				request.Method = "post"
			}
		case "-G", "--get":
			request.Method = "get"
			bodyForParams = true
		case "--url":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return shellCurlRequest{}, parseError("%s requires a value", token)
			}
			rawURL = value
		case "--compressed", "-i", "--include", "-s", "--silent", "-v", "--verbose", "-k", "--insecure":
			// Accepted DevTools/browser noise flags. They do not change goscrapling's
			// hermetic shell behavior in this bounded command seam.
		default:
			if strings.HasPrefix(token, "-") {
				return shellCurlRequest{}, parseError("unsupported curl option %q", token)
			}
			if rawURL != "" {
				return shellCurlRequest{}, parseError("curl command has multiple URLs")
			}
			rawURL = token
		}
	}
	if rawURL == "" {
		return shellCurlRequest{}, parseError("curl command requires a URL")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return shellCurlRequest{}, parseError("curl command has invalid URL %q", rawURL)
	}
	for key, values := range parsed.Query() {
		for _, value := range values {
			request.Params.Add(key, value)
		}
	}
	parsed.RawQuery = ""
	request.URL = parsed.String()
	if bodyForParams && request.Body != "" {
		values, err := url.ParseQuery(request.Body)
		if err != nil {
			return shellCurlRequest{}, parseError("curl -G data must be query encoded")
		}
		for key, parsedValues := range values {
			for _, value := range parsedValues {
				request.Params.Add(key, value)
			}
		}
		request.Body = ""
	}
	return request, nil
}

func nextToken(tokens []string, index *int, name string) (string, bool) {
	if *index+1 >= len(tokens) || tokens[*index+1] == "" {
		return "", false
	}
	*index = *index + 1
	return tokens[*index], true
}

func mergeShellCookies(cookies map[string]string, raw string) {
	for _, part := range strings.Split(raw, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || name == "" {
			continue
		}
		cookies[name] = value
	}
}

func splitShellWords(command string) ([]string, error) {
	var words []string
	var current strings.Builder
	var quote rune
	escaped := false
	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case ' ', '\t', '\n', '\r':
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words, nil
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

func parseError(format string, args ...any) error {
	return diagnostics.ParseError(format, args...)
}

func nextArg(args []string, index *int, name string) (string, bool) {
	return arguments.NextValue(args, index, name)
}

func parseHeader(value string) (string, string, error) {
	return arguments.ParseHeader(value)
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
