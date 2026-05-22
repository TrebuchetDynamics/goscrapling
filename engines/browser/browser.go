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
	ExtraFlags       []string
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
	ExtraFlags       []string
}

type BrowserResult struct {
	URL        string
	StatusCode int
	Reason     string
	Headers    http.Header
	Body       []byte
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
		Store: opts.Store,
	})
}
