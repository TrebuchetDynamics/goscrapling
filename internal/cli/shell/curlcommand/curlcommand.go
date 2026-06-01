package curlcommand

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/cli/arguments"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
)

// Request is the shell command's parsed curl request contract.
type Request struct {
	Method  string
	URL     string
	Headers http.Header
	Cookies map[string]string
	Params  url.Values
	Body    string
}

// Parse converts the supported curl subset into a request without executing it.
func Parse(command string) (Request, error) {
	tokens, err := splitWords(command)
	if err != nil {
		return Request{}, parseError("curl command parse error: %v", err)
	}
	if len(tokens) > 0 && tokens[0] == "curl" {
		tokens = tokens[1:]
	}
	request := Request{
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
				return Request{}, parseError("%s requires a value", token)
			}
			request.Method = strings.ToLower(value)
		case "-H", "--header":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return Request{}, parseError("%s requires a value", token)
			}
			key, headerValue, err := arguments.ParseHeader(value)
			if err != nil {
				return Request{}, err
			}
			if strings.EqualFold(key, "Cookie") {
				mergeCookies(request.Cookies, headerValue)
			} else {
				request.Headers.Add(key, headerValue)
			}
		case "-b", "--cookie":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return Request{}, parseError("%s requires a value", token)
			}
			mergeCookies(request.Cookies, value)
		case "-d", "--data", "--data-raw", "--data-binary":
			value, ok := nextToken(tokens, &i, token)
			if !ok {
				return Request{}, parseError("%s requires a value", token)
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
				return Request{}, parseError("%s requires a value", token)
			}
			rawURL = value
		case "--compressed", "-i", "--include", "-s", "--silent", "-v", "--verbose", "-k", "--insecure":
			// Accepted DevTools/browser noise flags. They do not change goscrapling's
			// hermetic shell behavior in this bounded command seam.
		default:
			if strings.HasPrefix(token, "-") {
				return Request{}, parseError("unsupported curl option %q", token)
			}
			if rawURL != "" {
				return Request{}, parseError("curl command has multiple URLs")
			}
			rawURL = token
		}
	}
	if rawURL == "" {
		return Request{}, parseError("curl command requires a URL")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return Request{}, parseError("curl command has invalid URL %q", rawURL)
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
			return Request{}, parseError("curl -G data must be query encoded")
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

func mergeCookies(cookies map[string]string, raw string) {
	for _, part := range strings.Split(raw, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || name == "" {
			continue
		}
		cookies[name] = value
	}
}

func splitWords(command string) ([]string, error) {
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

func parseError(format string, args ...any) error {
	return diagnostics.ParseError(format, args...)
}
