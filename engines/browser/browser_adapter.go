package browser

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	cdppage "github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type ChromedpBrowserOptions struct {
	ExecutablePath string
}

type ChromedpBrowserEngine struct {
	options ChromedpBrowserOptions
}

func NewChromedpBrowserEngine(opts ChromedpBrowserOptions) *ChromedpBrowserEngine {
	return &ChromedpBrowserEngine{options: opts}
}

func (e *ChromedpBrowserEngine) Fetch(ctx context.Context, request BrowserRequest) (BrowserResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}

	allocatorCtx, allocatorCancel := chromedpAllocatorContext(ctx, e, request)
	defer allocatorCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx)
	defer browserCancel()

	documentCapture := &chromedpDocumentCapture{}
	documentCapture.listen(browserCtx)
	xhrCapture, err := newChromedpXHRCapture(request.CaptureXHR)
	if err != nil {
		return BrowserResult{}, err
	}
	if xhrCapture != nil {
		xhrCapture.listen(browserCtx)
	}

	var renderedHTML string
	var screenshot []byte
	actions := []chromedp.Action{network.Enable()}
	if blockedPatterns := browserBlockedURLPatterns(request); len(blockedPatterns) > 0 {
		actions = append(actions, network.SetBlockedURLs().WithURLPatterns(chromedpBlockPatterns(blockedPatterns)))
	}
	if request.UserAgent != "" {
		userAgent := emulation.SetUserAgentOverride(request.UserAgent)
		if request.Locale != "" {
			userAgent = userAgent.WithAcceptLanguage(request.Locale)
		}
		actions = append(actions, userAgent)
	}
	if len(request.Headers) > 0 {
		actions = append(actions, network.SetExtraHTTPHeaders(chromedpNetworkHeaders(request.Headers)))
	}
	if request.TimezoneID != "" {
		actions = append(actions, emulation.SetTimezoneOverride(request.TimezoneID))
	}
	if script := browserStealthInitScript(request); script != "" {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := cdppage.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
			return err
		}))
	}
	actions = append(actions, chromedpCookieActions(request.Cookies, request.URL)...)
	actions = append(actions, chromedp.Navigate(request.URL))
	if request.LoadDOM {
		actions = append(actions, chromedp.WaitReady("body", chromedp.ByQuery))
	}
	actions = append(actions, chromedpActions(request.Actions)...)
	if request.WaitSelector.Selector != "" {
		actions = append(actions, chromedpWaitSelector(request.WaitSelector))
	}
	if request.NetworkIdle {
		actions = append(actions, chromedp.Sleep(500*time.Millisecond))
	}
	if request.Wait > 0 {
		actions = append(actions, chromedp.Sleep(request.Wait))
	}
	if request.Screenshot.Enabled {
		actions = append(actions, chromedpScreenshotAction(request.Screenshot, &screenshot))
	}
	actions = append(actions, chromedp.OuterHTML("html", &renderedHTML, chromedp.ByQuery))

	response, err := chromedp.RunResponse(browserCtx, actions...)
	if err != nil {
		return BrowserResult{}, err
	}

	resultURL := request.URL
	statusCode := http.StatusOK
	reason := http.StatusText(http.StatusOK)
	headers := http.Header{}
	if response != nil {
		if response.URL != "" {
			resultURL = response.URL
		}
		if response.Status > 0 {
			statusCode = int(response.Status)
		}
		if response.StatusText != "" {
			reason = response.StatusText
		} else if statusText := http.StatusText(statusCode); statusText != "" {
			reason = statusText
		}
		headers = httpHeadersFromChromedp(response.Headers)
		if headers.Get("Content-Type") == "" && response.MimeType != "" {
			headers.Set("Content-Type", response.MimeType)
		}
	}

	body := []byte(renderedHTML)
	if !browserResponseIsHTML(headers, response) {
		if rawBody, ok := documentCapture.body(browserCtx); ok {
			body = rawBody
		}
	}

	var capturedXHR []BrowserResult
	if xhrCapture != nil {
		capturedXHR = xhrCapture.results(browserCtx)
	}

	return BrowserResult{
		URL:         resultURL,
		StatusCode:  statusCode,
		Reason:      reason,
		Headers:     headers,
		Body:        body,
		Screenshot:  screenshot,
		CapturedXHR: capturedXHR,
	}, nil
}

type chromedpCapturedXHR struct {
	requestID network.RequestID
	response  *network.Response
}

type chromedpDocumentCapture struct {
	mu   sync.Mutex
	item chromedpCapturedXHR
}

func (c *chromedpDocumentCapture) listen(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {
		received, ok := ev.(*network.EventResponseReceived)
		if !ok || received.Response == nil || received.Type != network.ResourceTypeDocument {
			return
		}
		c.mu.Lock()
		c.item = chromedpCapturedXHR{requestID: received.RequestID, response: received.Response}
		c.mu.Unlock()
	})
}

func (c *chromedpDocumentCapture) body(ctx context.Context) ([]byte, bool) {
	c.mu.Lock()
	item := c.item
	c.mu.Unlock()
	if item.requestID == "" {
		return nil, false
	}
	body, err := network.GetResponseBody(item.requestID).Do(ctx)
	if err != nil {
		return nil, false
	}
	return body, true
}

type chromedpXHRCapture struct {
	pattern *regexp.Regexp
	mu      sync.Mutex
	items   []chromedpCapturedXHR
}

func newChromedpXHRCapture(pattern string) (*chromedpXHRCapture, error) {
	if pattern == "" {
		return nil, nil
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid capture_xhr pattern: %v", ErrBrowserOptions, err)
	}
	return &chromedpXHRCapture{pattern: compiled}, nil
}

func (c *chromedpXHRCapture) listen(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {
		if c == nil {
			return
		}
		received, ok := ev.(*network.EventResponseReceived)
		if !ok || received.Response == nil {
			return
		}
		if received.Type != network.ResourceTypeXHR && received.Type != network.ResourceTypeFetch {
			return
		}
		if !c.pattern.MatchString(received.Response.URL) {
			return
		}

		c.mu.Lock()
		c.items = append(c.items, chromedpCapturedXHR{requestID: received.RequestID, response: received.Response})
		c.mu.Unlock()
	})
}

func (c *chromedpXHRCapture) results(ctx context.Context) []BrowserResult {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	items := append([]chromedpCapturedXHR(nil), c.items...)
	c.mu.Unlock()
	if len(items) == 0 {
		return nil
	}

	results := make([]BrowserResult, 0, len(items))
	for _, item := range items {
		results = append(results, browserResultFromChromedpResponse(ctx, item))
	}
	return results
}

func browserResponseIsHTML(headers http.Header, response *network.Response) bool {
	contentType := strings.ToLower(headers.Get("Content-Type"))
	if strings.Contains(contentType, "html") {
		return true
	}
	if response != nil && strings.Contains(strings.ToLower(response.MimeType), "html") {
		return true
	}
	return contentType == "" && (response == nil || response.MimeType == "")
}

func browserResultFromChromedpResponse(ctx context.Context, item chromedpCapturedXHR) BrowserResult {
	response := item.response
	body, _ := network.GetResponseBody(item.requestID).Do(ctx)
	statusCode := http.StatusOK
	reason := http.StatusText(http.StatusOK)
	headers := http.Header{}
	resultURL := ""
	if response != nil {
		resultURL = response.URL
		if response.Status > 0 {
			statusCode = int(response.Status)
		}
		if response.StatusText != "" {
			reason = response.StatusText
		} else if statusText := http.StatusText(statusCode); statusText != "" {
			reason = statusText
		}
		headers = httpHeadersFromChromedp(response.Headers)
		if headers.Get("Content-Type") == "" && response.MimeType != "" {
			headers.Set("Content-Type", response.MimeType)
		}
	}
	return BrowserResult{
		URL:        resultURL,
		StatusCode: statusCode,
		Reason:     reason,
		Headers:    headers,
		Body:       body,
	}
}

func chromedpScreenshotAction(screenshot BrowserScreenshotOptions, output *[]byte) chromedp.Action {
	if screenshot.Selector != "" {
		return chromedp.Screenshot(screenshot.Selector, output, chromedp.ByQuery)
	}
	if screenshot.FullPage {
		quality := screenshot.Quality
		if quality == 0 {
			quality = 100
		}
		return chromedp.FullScreenshot(output, quality)
	}
	return chromedp.CaptureScreenshot(output)
}

func chromedpAllocatorContext(ctx context.Context, engine *ChromedpBrowserEngine, request BrowserRequest) (context.Context, context.CancelFunc) {
	if request.CDPURL != "" {
		return chromedp.NewRemoteAllocator(ctx, request.CDPURL)
	}

	allocOptions := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	allocOptions = append(allocOptions,
		chromedp.Flag("headless", request.Headless),
		chromedp.NoSandbox,
	)
	if request.Proxy.Server != "" {
		allocOptions = append(allocOptions, chromedp.ProxyServer(request.Proxy.Server))
	}
	if request.RealChrome && (engine == nil || engine.options.ExecutablePath == "") {
		allocOptions = append(allocOptions, chromedp.ExecPath("google-chrome"))
	}
	if engine != nil && engine.options.ExecutablePath != "" {
		allocOptions = append(allocOptions, chromedp.ExecPath(engine.options.ExecutablePath))
	}
	if request.DNSOverHTTPS {
		allocOptions = append(allocOptions,
			chromedp.Flag("enable-features", "DnsOverHttps"),
			chromedp.Flag("dns-over-https-templates", "https://cloudflare-dns.com/dns-query"),
		)
	}
	for _, flag := range request.ExtraFlags {
		allocOptions = append(allocOptions, chromedpFlag(flag))
	}
	return chromedp.NewExecAllocator(ctx, allocOptions...)
}

func chromedpFlag(flag string) chromedp.ExecAllocatorOption {
	flag = strings.TrimSpace(strings.TrimPrefix(flag, "--"))
	name, value, hasValue := strings.Cut(flag, "=")
	if hasValue {
		return chromedp.Flag(name, value)
	}
	return chromedp.Flag(name, true)
}

func chromedpCookieActions(cookies []*http.Cookie, rawURL string) []chromedp.Action {
	output := make([]chromedp.Action, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		action := network.SetCookie(cookie.Name, cookie.Value).WithURL(rawURL)
		if cookie.Domain != "" {
			action = action.WithDomain(cookie.Domain)
		}
		if cookie.Path != "" {
			action = action.WithPath(cookie.Path)
		}
		if cookie.Secure {
			action = action.WithSecure(true)
		}
		if cookie.HttpOnly {
			action = action.WithHTTPOnly(true)
		}
		output = append(output, action)
	}
	return output
}

func chromedpBlockPatterns(patterns []string) []*network.BlockPattern {
	output := make([]*network.BlockPattern, 0, len(patterns))
	for _, pattern := range patterns {
		output = append(output, &network.BlockPattern{URLPattern: pattern, Block: true})
	}
	return output
}

func chromedpActions(actions []BrowserAction) []chromedp.Action {
	output := make([]chromedp.Action, 0, len(actions))
	for _, action := range actions {
		switch action.Kind {
		case BrowserActionClick:
			output = append(output, chromedp.Click(action.Selector, chromedp.ByQuery))
		case BrowserActionFill:
			output = append(output, chromedp.SetValue(action.Selector, action.Value, chromedp.ByQuery))
		case BrowserActionWaitForSelector:
			output = append(output, chromedp.WaitReady(action.Selector, chromedp.ByQuery))
		case BrowserActionEvaluate:
			output = append(output, chromedp.Evaluate(action.Value, nil))
		}
	}
	return output
}

func chromedpWaitSelector(wait BrowserWaitSelector) chromedp.Action {
	switch wait.State {
	case BrowserWaitDetached:
		return chromedp.WaitNotPresent(wait.Selector, chromedp.ByQuery)
	case BrowserWaitHidden:
		return chromedp.WaitNotVisible(wait.Selector, chromedp.ByQuery)
	case BrowserWaitVisible:
		return chromedp.WaitVisible(wait.Selector, chromedp.ByQuery)
	default:
		return chromedp.WaitReady(wait.Selector, chromedp.ByQuery)
	}
}

func chromedpNetworkHeaders(headers http.Header) network.Headers {
	output := network.Headers{}
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		output[key] = strings.Join(values, ", ")
	}
	return output
}

func httpHeadersFromChromedp(headers network.Headers) http.Header {
	output := http.Header{}
	for key, value := range headers {
		switch typed := value.(type) {
		case string:
			output.Add(key, typed)
		case []string:
			for _, item := range typed {
				output.Add(key, item)
			}
		case []any:
			for _, item := range typed {
				output.Add(key, fmt.Sprint(item))
			}
		default:
			output.Add(key, fmt.Sprint(value))
		}
	}
	return output
}
