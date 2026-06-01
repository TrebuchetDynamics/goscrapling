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
	if len(tokens) > 0 && tokens[0].text == "curl" {
		tokens = tokens[1:]
	}
	request := Request{
		Method:  "get",
		Headers: http.Header{},
		Cookies: map[string]string{},
		Params:  url.Values{},
	}
	var rawURL string
	var data curlData
	explicitMethod := false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token.text {
		case "-X", "--request":
			value, ok := nextToken(tokens, &i, token.text)
			if !ok {
				return Request{}, parseError("%s requires a value", token.text)
			}
			request.Method = strings.ToLower(value.text)
			explicitMethod = true
		case "-H", "--header":
			value, ok := nextToken(tokens, &i, token.text)
			if !ok {
				return Request{}, parseError("%s requires a value", token.text)
			}
			key, headerValue, err := arguments.ParseHeader(value.text)
			if err != nil {
				return Request{}, err
			}
			if strings.EqualFold(key, "Cookie") {
				mergeCookies(request.Cookies, headerValue)
			} else {
				request.Headers.Add(key, headerValue)
			}
		case "-b", "--cookie":
			value, ok := nextToken(tokens, &i, token.text)
			if !ok {
				return Request{}, parseError("%s requires a value", token.text)
			}
			mergeCookies(request.Cookies, value.text)
		case "-d", "--data", "--data-raw", "--data-binary":
			value, ok := nextToken(tokens, &i, token.text)
			if !ok {
				return Request{}, parseError("%s requires a value", token.text)
			}
			data.add(value)
		case "-G", "--get":
			data.forceQueryParams = true
		case "--url":
			value, ok := nextToken(tokens, &i, token.text)
			if !ok {
				return Request{}, parseError("%s requires a value", token.text)
			}
			rawURL = value.text
		case "--compressed", "-i", "--include", "-s", "--silent", "-v", "--verbose", "-k", "--insecure":
			// Accepted DevTools/browser noise flags. They do not change goscrapling's
			// hermetic shell behavior in this bounded command seam.
		default:
			if strings.HasPrefix(token.text, "-") {
				return Request{}, parseError("unsupported curl option %q", token.text)
			}
			if rawURL != "" {
				return Request{}, parseError("curl command has multiple URLs")
			}
			rawURL = token.text
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
	if err := data.applyTo(&request, explicitMethod); err != nil {
		return Request{}, err
	}
	return request, nil
}

type curlData struct {
	parts            []string
	forceQueryParams bool
}

func (d *curlData) add(value curlWord) {
	d.parts = append(d.parts, value.text)
}

func (d curlData) applyTo(request *Request, explicitMethod bool) error {
	if len(d.parts) == 0 {
		return nil
	}
	body := strings.Join(d.parts, "&")
	if d.forceQueryParams {
		request.Method = "get"
		values, err := url.ParseQuery(body)
		if err != nil {
			return parseError("curl -G data must be query encoded")
		}
		for key, parsedValues := range values {
			for _, value := range parsedValues {
				request.Params.Add(key, value)
			}
		}
		request.Body = ""
		return nil
	}
	request.Body = body
	request.Method = methodWithBodyData(request.Method, explicitMethod)
	return nil
}

func methodWithBodyData(currentMethod string, explicitMethod bool) string {
	if explicitMethod {
		return currentMethod
	}
	if currentMethod == "get" {
		return "post"
	}
	return currentMethod
}

func nextToken(tokens []curlWord, index *int, name string) (curlWord, bool) {
	if *index+1 >= len(tokens) {
		return curlWord{}, false
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

type curlWord struct {
	text string
}

func splitWords(command string) ([]curlWord, error) {
	var words []curlWord
	var current strings.Builder
	var quote rune
	inWord := false
	escaped := false
	for _, r := range command {
		if escaped {
			if appendEscapedCurlRune(&current, r) {
				inWord = true
			}
			escaped = false
			continue
		}
		if shouldStartCurlEscape(quote, r) {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
				inWord = true
			}
			continue
		}
		switch r {
		case '\'', '"':
			if r == '\'' && strings.HasSuffix(current.String(), "$") {
				trimBuilderSuffix(&current, "$")
			}
			quote = r
			inWord = true
		case ' ', '\t', '\n', '\r':
			if inWord {
				words = append(words, curlWord{text: current.String()})
				current.Reset()
				inWord = false
			}
		default:
			current.WriteRune(r)
			inWord = true
		}
	}
	if escaped {
		current.WriteRune('\\')
		inWord = true
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	if inWord {
		words = append(words, curlWord{text: current.String()})
	}
	return words, nil
}

func shouldStartCurlEscape(quote rune, r rune) bool {
	if r != '\\' {
		return false
	}
	return quote != '\''
}

func appendEscapedCurlRune(current *strings.Builder, r rune) bool {
	if isShellLineContinuationRune(r) {
		return false
	}
	current.WriteRune(r)
	return true
}

func isShellLineContinuationRune(r rune) bool {
	return r == '\n' || r == '\r'
}

func trimBuilderSuffix(builder *strings.Builder, suffix string) {
	trimmed := strings.TrimSuffix(builder.String(), suffix)
	builder.Reset()
	builder.WriteString(trimmed)
}

func parseError(format string, args ...any) error {
	return diagnostics.ParseError(format, args...)
}
