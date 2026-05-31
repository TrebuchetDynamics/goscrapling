package cli

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"golang.org/x/net/html"
)

type extractPlan struct {
	method      string
	rawURL      string
	outputPath  string
	selector    string
	aiTargeted  bool
	useBrowser  bool
	request     goscrapling.RequestOptions
	browserOpts browser.BrowserOptions
}

type extractOptions struct {
	cssSelector     string
	headers         http.Header
	queryParams     url.Values
	body            []byte
	timeout         time.Duration
	followRedirects goscrapling.RedirectPolicy
	aiTargeted      bool
}

type browserExtractOptions struct {
	cssSelector string
	aiTargeted  bool
	browserOpts browser.BrowserOptions
}

var fetchBrowserExtract = defaultFetchBrowserExtract

func runExtract(stdout io.Writer, args []string) error {
	if len(args) == 0 {
		return parseError("missing extract command")
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, err := fmt.Fprintln(stdout, usage)
		return err
	}

	plan, err := parseExtractPlan(args)
	if err != nil {
		return err
	}
	return executeExtractPlan(stdout, plan)
}

func parseExtractPlan(args []string) (extractPlan, error) {
	method := strings.ToLower(args[0])
	if isStaticExtractMethod(method) {
		rawURL, outputPath, opts, err := parseExtractArgs(method, args[1:])
		if err != nil {
			return extractPlan{}, err
		}
		rawURL, err = appendQueryParams(rawURL, opts.queryParams)
		if err != nil {
			return extractPlan{}, err
		}

		return extractPlan{
			method:     method,
			rawURL:     rawURL,
			outputPath: outputPath,
			selector:   opts.cssSelector,
			aiTargeted: opts.aiTargeted,
			request: goscrapling.RequestOptions{
				Headers:         opts.headers,
				Body:            bytes.NewReader(opts.body),
				Timeout:         opts.timeout,
				FollowRedirects: opts.followRedirects,
			},
		}, nil
	}
	if isBrowserExtractMethod(method) {
		rawURL, outputPath, opts, err := parseBrowserExtractArgs(method, args[1:])
		if err != nil {
			return extractPlan{}, err
		}
		return extractPlan{
			method:      method,
			rawURL:      rawURL,
			outputPath:  outputPath,
			selector:    opts.cssSelector,
			aiTargeted:  opts.aiTargeted,
			useBrowser:  true,
			browserOpts: opts.browserOpts,
		}, nil
	}
	return extractPlan{}, parseError("unknown extract command %q", args[0])
}

func executeExtractPlan(stdout io.Writer, plan extractPlan) error {
	response, err := executeExtractFetch(plan)
	if err != nil {
		return fmt.Errorf("extract %s %q: %w", plan.method, plan.rawURL, err)
	}
	if plan.aiTargeted {
		response, err = aiTargetedResponse(response)
		if err != nil {
			return err
		}
	}

	body, err := renderExtractOutput(response, plan.outputPath, plan.selector)
	if err != nil {
		return err
	}
	if err := writeOutputFile(plan.outputPath, body); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "wrote %s\n", plan.outputPath)
	return err
}

func executeExtractFetch(plan extractPlan) (*goscrapling.Response, error) {
	if plan.useBrowser {
		opts := plan.browserOpts
		if plan.aiTargeted {
			opts.BlockAds = true
		}
		return fetchBrowserExtract(context.Background(), plan.rawURL, opts)
	}
	return fetchStatic(plan.method, plan.rawURL, plan.request)
}

func isStaticExtractMethod(method string) bool {
	switch method {
	case "get", "post", "put", "delete":
		return true
	default:
		return false
	}
}

func isBrowserExtractMethod(method string) bool {
	switch method {
	case "fetch", "stealthy-fetch":
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

func defaultFetchBrowserExtract(ctx context.Context, rawURL string, opts browser.BrowserOptions) (*goscrapling.Response, error) {
	engine := browser.NewChromedpBrowserEngine(browser.ChromedpBrowserOptions{})
	fetcher := browser.BrowserFetcher{Engine: engine}
	return fetcher.Fetch(ctx, rawURL, opts)
}

func parseExtractArgs(method string, args []string) (string, string, extractOptions, error) {
	opts := extractOptions{
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
		case "--ai-targeted":
			opts.aiTargeted = true
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

func parseBrowserExtractArgs(method string, args []string) (string, string, browserExtractOptions, error) {
	opts := browserExtractOptions{browserOpts: browser.BrowserOptions{
		Headers:  http.Header{},
		Headless: true,
		LoadDOM:  true,
		Timeout:  30 * time.Second,
	}}
	if method == "stealthy-fetch" {
		opts.browserOpts.Stealth = browser.BrowserStealthOptions{Enabled: true, GenerateHeaders: true}
	}

	positionals := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--css-selector", "-s":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			opts.cssSelector = value
		case "--extra-headers", "--headers", "-H":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			key, headerValue, err := parseHeader(value)
			if err != nil {
				return "", "", opts, err
			}
			opts.browserOpts.Headers.Add(key, headerValue)
		case "--headless":
			opts.browserOpts.Headless = true
		case "--no-headless":
			opts.browserOpts.Headless = false
		case "--disable-resources":
			opts.browserOpts.DisableResources = true
		case "--enable-resources":
			opts.browserOpts.DisableResources = false
		case "--network-idle":
			opts.browserOpts.NetworkIdle = true
		case "--no-network-idle":
			opts.browserOpts.NetworkIdle = false
		case "--timeout":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			milliseconds, err := strconv.Atoi(value)
			if err != nil || milliseconds <= 0 {
				return "", "", opts, parseError("timeout must be a positive integer")
			}
			opts.browserOpts.Timeout = time.Duration(milliseconds) * time.Millisecond
		case "--wait":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			milliseconds, err := strconv.Atoi(value)
			if err != nil || milliseconds < 0 {
				return "", "", opts, parseError("wait must be a non-negative integer")
			}
			opts.browserOpts.Wait = time.Duration(milliseconds) * time.Millisecond
		case "--wait-selector":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			opts.browserOpts.WaitSelector = browser.BrowserWaitSelector{Selector: value}
		case "--locale":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			opts.browserOpts.Locale = value
		case "--real-chrome":
			opts.browserOpts.RealChrome = true
		case "--no-real-chrome":
			opts.browserOpts.RealChrome = false
		case "--proxy":
			value, ok := nextArg(args, &i, arg)
			if !ok {
				return "", "", opts, parseError("%s requires a value", arg)
			}
			opts.browserOpts.Proxy.Server = value
		case "--dns-over-https":
			opts.browserOpts.DNSOverHTTPS = true
		case "--no-dns-over-https":
			opts.browserOpts.DNSOverHTTPS = false
		case "--block-ads":
			opts.browserOpts.BlockAds = true
		case "--no-block-ads":
			opts.browserOpts.BlockAds = false
		case "--ai-targeted":
			opts.aiTargeted = true
		case "--block-webrtc":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.BlockWebRTC = true
		case "--allow-webrtc":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.BlockWebRTC = false
		case "--solve-cloudflare":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.SolveCloudflare = true
		case "--no-solve-cloudflare":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.SolveCloudflare = false
		case "--allow-webgl":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.DisableWebGL = false
		case "--block-webgl":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.DisableWebGL = true
		case "--hide-canvas":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.HideCanvas = true
		case "--show-canvas":
			if method != "stealthy-fetch" {
				return "", "", opts, parseError("unknown option %q", arg)
			}
			opts.browserOpts.Stealth.HideCanvas = false
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
	case ".md":
		htmlBody := response.Body()
		if selector != "" {
			selectedHTML, err := response.CSS(selector).HTML()
			if err != nil {
				return nil, err
			}
			htmlBody = []byte(selectedHTML)
		}
		markdown, err := browser.HTMLToMarkdown(htmlBody)
		if err != nil {
			return nil, err
		}
		return []byte(markdown), nil
	case ".txt", "":
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

func aiTargetedResponse(response *goscrapling.Response) (*goscrapling.Response, error) {
	cleaned, err := aiTargetedHTML(response.Body())
	if err != nil {
		return nil, err
	}
	return goscrapling.NewResponse(bytes.NewReader(cleaned), goscrapling.ResponseOptions{
		URL:         response.URL(),
		StatusCode:  response.StatusCode(),
		Reason:      response.Reason(),
		Headers:     response.Headers(),
		Request:     response.Request(),
		Encoding:    response.Encoding(),
		Cookies:     response.Cookies(),
		History:     response.History(),
		Meta:        response.Meta(),
		CapturedXHR: response.CapturedXHR(),
	})
}

func aiTargetedHTML(body []byte) ([]byte, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	stripAITargetedNoise(root)
	target := firstElementByTag(root, "main")
	if target == nil {
		target = firstElementByTag(root, "body")
	}
	if target == nil {
		target = root
	}
	var cleaned bytes.Buffer
	if err := html.Render(&cleaned, target); err != nil {
		return nil, err
	}
	return cleaned.Bytes(), nil
}

func stripAITargetedNoise(node *html.Node) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if shouldRemoveAITargetedNode(child) {
			node.RemoveChild(child)
		} else {
			if child.Type == html.TextNode {
				child.Data = removeZeroWidthRunes(child.Data)
			}
			stripAITargetedNoise(child)
		}
		child = next
	}
}

func shouldRemoveAITargetedNode(node *html.Node) bool {
	if node.Type == html.CommentNode {
		return true
	}
	if node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "script", "style", "noscript", "svg", "template", "nav", "header", "footer", "aside":
		return true
	}
	if _, ok := htmlAttr(node, "hidden"); ok {
		return true
	}
	if value, ok := htmlAttr(node, "aria-hidden"); ok && strings.EqualFold(strings.TrimSpace(value), "true") {
		return true
	}
	if strings.EqualFold(node.Data, "input") {
		if value, ok := htmlAttr(node, "type"); ok && strings.EqualFold(strings.TrimSpace(value), "hidden") {
			return true
		}
	}
	if style, ok := htmlAttr(node, "style"); ok {
		normalized := strings.ToLower(strings.ReplaceAll(style, " ", ""))
		if strings.Contains(normalized, "display:none") || strings.Contains(normalized, "visibility:hidden") {
			return true
		}
	}
	return false
}

func firstElementByTag(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, tag) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := firstElementByTag(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func htmlAttr(node *html.Node, name string) (string, bool) {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val, true
		}
	}
	return "", false
}

func removeZeroWidthRunes(value string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u200b', '\u200c', '\u200d', '\ufeff', '\u2060':
			return -1
		default:
			return r
		}
	}, value)
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
