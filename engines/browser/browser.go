package browser

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"
)

var ErrMissingBrowserEngine = errors.New("missing browser engine")

type BrowserEngine interface {
	Fetch(ctx context.Context, request BrowserRequest) (BrowserResult, error)
}

type BrowserFetcher struct {
	Engine BrowserEngine
}

type BrowserOptions struct {
	Headers          http.Header
	UserAgent        string
	Cookies          []*http.Cookie
	Locale           string
	TimezoneID       string
	Proxy            BrowserProxyOptions
	CDPURL           string
	RealChrome       bool
	Headless         bool
	DisableResources bool
	BlockedDomains   []string
	BlockAds         bool
	DNSOverHTTPS     bool
	NetworkIdle      bool
	LoadDOM          bool
	Timeout          time.Duration
	Wait             time.Duration
	WaitSelector     BrowserWaitSelector
	Actions          []BrowserAction
	CaptureXHR       string
	Screenshot       BrowserScreenshotOptions
	ExtraFlags       []string
	Stealth          BrowserStealthOptions
	Store            Store
}

type BrowserRequest struct {
	URL              string
	Headers          http.Header
	UserAgent        string
	Cookies          []*http.Cookie
	Locale           string
	TimezoneID       string
	Proxy            BrowserProxyOptions
	CDPURL           string
	RealChrome       bool
	Headless         bool
	DisableResources bool
	BlockedDomains   []string
	BlockAds         bool
	DNSOverHTTPS     bool
	NetworkIdle      bool
	LoadDOM          bool
	Timeout          time.Duration
	Wait             time.Duration
	WaitSelector     BrowserWaitSelector
	Actions          []BrowserAction
	CaptureXHR       string
	Screenshot       BrowserScreenshotOptions
	ExtraFlags       []string
	Stealth          BrowserStealthOptions
}

type BrowserResult struct {
	URL         string
	StatusCode  int
	Reason      string
	Headers     http.Header
	Body        []byte
	Screenshot  []byte
	CapturedXHR []BrowserResult
}

type BrowserWaitState string

const (
	BrowserWaitAttached BrowserWaitState = "attached"
	BrowserWaitDetached BrowserWaitState = "detached"
	BrowserWaitVisible  BrowserWaitState = "visible"
	BrowserWaitHidden   BrowserWaitState = "hidden"
)

type BrowserWaitSelector struct {
	Selector string
	State    BrowserWaitState
}

type BrowserActionKind string

const (
	BrowserActionClick           BrowserActionKind = "click"
	BrowserActionFill            BrowserActionKind = "fill"
	BrowserActionWaitForSelector BrowserActionKind = "wait_for_selector"
	BrowserActionEvaluate        BrowserActionKind = "evaluate"
)

type BrowserAction struct {
	Kind     BrowserActionKind
	Selector string
	Value    string
}

type BrowserScreenshotOptions struct {
	Enabled  bool
	FullPage bool
	Selector string
	Quality  int
}

func (f BrowserFetcher) Fetch(ctx context.Context, rawURL string, opts BrowserOptions) (*Response, error) {
	if f.Engine == nil {
		return nil, ErrMissingBrowserEngine
	}
	if ctx == nil {
		ctx = context.Background()
	}

	request, err := newBrowserRequest(rawURL, opts)
	if err != nil {
		return nil, err
	}

	result, err := f.Engine.Fetch(ctx, request)
	if err != nil {
		return nil, err
	}

	responseURL := result.URL
	if responseURL == "" {
		responseURL = rawURL
	}
	capturedXHR, err := browserResponsesFromResults(result.CapturedXHR, opts.Store)
	if err != nil {
		return nil, err
	}
	return NewResponse(bytes.NewReader(result.Body), ResponseOptions{
		URL:        responseURL,
		StatusCode: result.StatusCode,
		Reason:     result.Reason,
		Headers:    result.Headers,
		Request: RequestMetadata{
			Method:  http.MethodGet,
			URL:     rawURL,
			Headers: request.Headers,
		},
		Meta:        browserResultMeta(result),
		CapturedXHR: capturedXHR,
		Store:       opts.Store,
	})
}

func browserResultMeta(result BrowserResult) map[string]any {
	if len(result.Screenshot) == 0 {
		return nil
	}
	return map[string]any{"screenshot": append([]byte(nil), result.Screenshot...)}
}

func browserResponsesFromResults(results []BrowserResult, store Store) ([]*Response, error) {
	if len(results) == 0 {
		return nil, nil
	}
	responses := make([]*Response, 0, len(results))
	for _, result := range results {
		response, err := NewResponse(bytes.NewReader(result.Body), ResponseOptions{
			URL:        result.URL,
			StatusCode: result.StatusCode,
			Reason:     result.Reason,
			Headers:    result.Headers,
			Request: RequestMetadata{
				Method: http.MethodGet,
				URL:    result.URL,
			},
			Meta:  browserResultMeta(result),
			Store: store,
		})
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}
