package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/goscrapling"
)

const usage = "usage: goscrapling extract {get|post|put|delete} <url> <output_file> [--css-selector <selector>] [-H <key: value>] [--timeout <seconds>] [-p <key=value>] [--data <body>] [--json <json>] [--no-follow-redirects]"

var ErrParse = errors.New("parse error")

type extractGetOptions struct {
	cssSelector     string
	headers         http.Header
	queryParams     url.Values
	body            []byte
	timeout         time.Duration
	followRedirects goscrapling.RedirectPolicy
}

func Run(stdout, stderr io.Writer, args []string) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 {
		return parseError("missing command")
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, err := fmt.Fprintln(stdout, usage)
		return err
	}
	if args[0] != "extract" {
		return parseError("unknown command %q", args[0])
	}
	return runExtract(stdout, args[1:])
}

func runExtract(stdout io.Writer, args []string) error {
	if len(args) == 0 {
		return parseError("missing extract command")
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, err := fmt.Fprintln(stdout, usage)
		return err
	}
	method := strings.ToLower(args[0])
	if !isStaticExtractMethod(method) {
		return parseError("unknown extract command %q", args[0])
	}

	rawURL, outputPath, opts, err := parseExtractGetArgs(method, args[1:])
	if err != nil {
		return err
	}
	rawURL, err = appendQueryParams(rawURL, opts.queryParams)
	if err != nil {
		return err
	}

	requestOptions := goscrapling.RequestOptions{
		Headers:         opts.headers,
		Body:            bytes.NewReader(opts.body),
		Timeout:         opts.timeout,
		FollowRedirects: opts.followRedirects,
	}
	response, err := fetchStatic(method, rawURL, requestOptions)
	if err != nil {
		return fmt.Errorf("extract %s %q: %w", method, rawURL, err)
	}

	body, err := renderExtractOutput(response, outputPath, opts.cssSelector)
	if err != nil {
		return err
	}
	if err := writeOutputFile(outputPath, body); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "wrote %s\n", outputPath)
	return err
}

func isStaticExtractMethod(method string) bool {
	switch method {
	case "get", "post", "put", "delete":
		return true
	default:
		return false
	}
}

func fetchStatic(method string, rawURL string, opts goscrapling.RequestOptions) (*goscrapling.Response, error) {
	fetcher := goscrapling.Fetcher{}
	switch method {
	case "get":
		return fetcher.Get(rawURL, opts)
	case "post":
		return fetcher.Post(rawURL, opts)
	case "put":
		return fetcher.Put(rawURL, opts)
	case "delete":
		return fetcher.Delete(rawURL, opts)
	default:
		return nil, parseError("unknown extract command %q", method)
	}
}

func parseExtractGetArgs(method string, args []string) (string, string, extractGetOptions, error) {
	opts := extractGetOptions{
		headers:         http.Header{},
		queryParams:     url.Values{},
		followRedirects: goscrapling.RedirectPolicySafe,
	}
	positionals := make([]string, 0, 2)
	var dataBody string
	var jsonBody string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--css-selector", "-s":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			opts.cssSelector = value
		case "--headers", "-H":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			key, headerValue, err := parseHeader(value)
			if err != nil {
				return "", "", opts, err
			}
			opts.headers.Add(key, headerValue)
		case "--params", "-p":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			key, paramValue, err := parseKeyValue(value, "query parameters")
			if err != nil {
				return "", "", opts, err
			}
			opts.queryParams.Add(key, paramValue)
		case "--data", "-d":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			dataBody = value
		case "--json", "-j":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			jsonBody = value
		case "--timeout":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			seconds, err := strconv.Atoi(value)
			if err != nil || seconds <= 0 {
				return "", "", opts, parseError("timeout must be a positive integer")
			}
			opts.timeout = time.Duration(seconds) * time.Second
		case "--no-follow-redirects":
			opts.followRedirects = goscrapling.RedirectPolicyNone
		case "--follow-redirects":
			opts.followRedirects = goscrapling.RedirectPolicySafe
		default:
			if strings.HasPrefix(arg, "-") {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 2 {
		return "", "", opts, parseError("extract %s requires url and output_file", method)
	}
	if dataBody != "" && jsonBody != "" {
		return "", "", opts, parseError("--data and --json cannot be used together")
	}
	if jsonBody != "" {
		if !json.Valid([]byte(jsonBody)) {
			return "", "", opts, parseError("invalid JSON body")
		}
		opts.body = []byte(jsonBody)
		if opts.headers.Get("Content-Type") == "" {
			opts.headers.Set("Content-Type", "application/json")
		}
	} else if dataBody != "" {
		opts.body = []byte(dataBody)
		if opts.headers.Get("Content-Type") == "" {
			opts.headers.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	return positionals[0], positionals[1], opts, nil
}

func nextArg(args []string, index *int, name string) (string, bool) {
	if *index+1 >= len(args) || args[*index+1] == "" {
		return "", false
	}
	*index = *index + 1
	return args[*index], true
}

func parseHeader(value string) (string, string, error) {
	key, headerValue, ok := strings.Cut(value, ":")
	key = strings.TrimSpace(key)
	headerValue = strings.TrimSpace(headerValue)
	if !ok || key == "" {
		return "", "", parseError("headers must use %q format", "Key: Value")
	}
	return key, headerValue, nil
}

func parseKeyValue(value string, name string) (string, string, error) {
	key, paramValue, ok := strings.Cut(value, "=")
	key = strings.TrimSpace(key)
	if !ok || key == "" {
		return "", "", parseError("%s must use %q format", name, "key=value")
	}
	return key, paramValue, nil
}

func appendQueryParams(rawURL string, params url.Values) (string, error) {
	if len(params) == 0 {
		return rawURL, nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	for key, values := range params {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func renderExtractOutput(response *goscrapling.Response, outputPath string, selector string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".html", ".htm":
		if selector == "" {
			return response.Body(), nil
		}
		selectedHTML, err := response.CSS(selector).HTML()
		if err != nil {
			return nil, err
		}
		return []byte(selectedHTML), nil
	case ".txt", ".md", "":
		return []byte(extractText(response, selector)), nil
	default:
		return nil, parseError("unsupported output extension %q", ext)
	}
}

func extractText(response *goscrapling.Response, selector string) string {
	if selector != "" {
		return response.CSS(selector).Text()
	}
	bodyText := response.CSS("body").Text()
	if bodyText != "" {
		return bodyText
	}
	return strings.TrimSpace(string(bytes.TrimSpace(response.Body())))
}

func writeOutputFile(path string, body []byte) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, body, 0o644)
}

func parseError(format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%w: %s\n%s", ErrParse, message, usage)
}
