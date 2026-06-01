package spiders

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
)

const (
	StaticRequestOptionsMetaKey  = "goscrapling:spiders:static_request_options"
	BrowserRequestOptionsMetaKey = "goscrapling:spiders:browser_request_options"
)

type StaticSessionAdapterOptions struct {
	Options fetchers.RequestOptions
}

type StaticSessionAdapter struct {
	session *fetchers.FetcherSession
	options fetchers.RequestOptions
}

func NewStaticSessionAdapter(session *fetchers.FetcherSession, opts StaticSessionAdapterOptions) *StaticSessionAdapter {
	return &StaticSessionAdapter{session: session, options: opts.Options}
}

func WithStaticRequestOptions(request Request, opts fetchers.RequestOptions) Request {
	request = request.clone()
	if request.Meta == nil {
		request.Meta = map[string]any{}
	}
	request.Meta[StaticRequestOptionsMetaKey] = opts
	return request
}

func (a *StaticSessionAdapter) Fetch(ctx context.Context, request Request) (*goscrapling.Response, error) {
	if a == nil || a.session == nil {
		return nil, fmt.Errorf("static session adapter requires a fetcher session")
	}
	opts := mergeStaticRequestOptions(a.options, staticRequestOptionsFromMeta(request.Meta))
	opts.Context = ctx
	opts.Headers = mergeHTTPHeaders(opts.Headers, request.Headers)
	if len(request.Body) > 0 {
		opts.Body = bytes.NewReader(request.Body)
	}
	applyStaticProxyMeta(&opts, request.Meta)

	switch request.MethodOrDefault() {
	case http.MethodGet:
		return a.session.Get(request.URL, opts)
	case http.MethodPost:
		return a.session.Post(request.URL, opts)
	case http.MethodPut:
		return a.session.Put(request.URL, opts)
	case http.MethodDelete:
		return a.session.Delete(request.URL, opts)
	default:
		return nil, fmt.Errorf("unsupported static spider method %q", request.MethodOrDefault())
	}
}

type BrowserSessionAdapterOptions struct {
	Options      browser.BrowserOptions
	ProxyRotator *fetchers.ProxyRotator
}

type BrowserSessionAdapter struct {
	session      *browser.BrowserSession
	options      browser.BrowserOptions
	proxyRotator *fetchers.ProxyRotator
	stealth      bool
}

func NewBrowserSessionAdapter(session *browser.BrowserSession, opts BrowserSessionAdapterOptions) *BrowserSessionAdapter {
	return &BrowserSessionAdapter{session: session, options: cloneSpiderBrowserOptions(opts.Options), proxyRotator: opts.ProxyRotator}
}

func NewStealthBrowserSessionAdapter(session *browser.BrowserSession, opts BrowserSessionAdapterOptions) *BrowserSessionAdapter {
	adapter := NewBrowserSessionAdapter(session, opts)
	adapter.stealth = true
	return adapter
}

func WithBrowserRequestOptions(request Request, opts browser.BrowserOptions) Request {
	request = request.clone()
	if request.Meta == nil {
		request.Meta = map[string]any{}
	}
	request.Meta[BrowserRequestOptionsMetaKey] = cloneSpiderBrowserOptions(opts)
	return request
}

func (a *BrowserSessionAdapter) Fetch(ctx context.Context, request Request) (*goscrapling.Response, error) {
	if a == nil || a.session == nil {
		return nil, fmt.Errorf("browser session adapter requires a browser session")
	}
	opts := mergeSpiderBrowserOptions(a.options, browserRequestOptionsFromMeta(request.Meta))
	opts.Headers = mergeHTTPHeaders(opts.Headers, request.Headers)
	if a.stealth {
		opts.Stealth.Enabled = true
	}

	selectedProxy := ""
	if proxy, ok := browserProxyFromMeta(request.Meta); ok {
		opts.Proxy = proxy
		selectedProxy = proxy.Server
	} else if opts.Proxy.Server == "" && a.proxyRotator != nil {
		rotated, err := a.proxyRotator.Next()
		if err != nil {
			return nil, err
		}
		opts.Proxy = browserProxyFromFetchProxy(rotated)
		selectedProxy = opts.Proxy.Server
	}

	response, err := a.session.Fetch(ctx, request.URL, opts)
	if err != nil || selectedProxy == "" {
		return response, err
	}
	return responseWithMergedMeta(response, map[string]any{"proxy": selectedProxy})
}

func (a *BrowserSessionAdapter) Close(context.Context) error {
	if a == nil || a.session == nil {
		return nil
	}
	return a.session.Close()
}

func staticRequestOptionsFromMeta(meta map[string]any) fetchers.RequestOptions {
	if meta == nil {
		return fetchers.RequestOptions{}
	}
	switch opts := meta[StaticRequestOptionsMetaKey].(type) {
	case fetchers.RequestOptions:
		return opts
	case *fetchers.RequestOptions:
		if opts != nil {
			return *opts
		}
	}
	return fetchers.RequestOptions{}
}

func browserRequestOptionsFromMeta(meta map[string]any) browser.BrowserOptions {
	if meta == nil {
		return browser.BrowserOptions{}
	}
	switch opts := meta[BrowserRequestOptionsMetaKey].(type) {
	case browser.BrowserOptions:
		return cloneSpiderBrowserOptions(opts)
	case *browser.BrowserOptions:
		if opts != nil {
			return cloneSpiderBrowserOptions(*opts)
		}
	}
	return browser.BrowserOptions{}
}

func mergeStaticRequestOptions(base, override fetchers.RequestOptions) fetchers.RequestOptions {
	merged := base
	merged.Headers = mergeHTTPHeaders(base.Headers, override.Headers)
	if override.Body != nil {
		merged.Body = override.Body
	}
	if len(override.Params) > 0 {
		merged.Params = override.Params
	}
	if len(override.Data) > 0 {
		merged.Data = override.Data
	}
	if override.JSON != nil {
		merged.JSON = override.JSON
	}
	if override.Auth != nil {
		merged.Auth = override.Auth
	}
	if override.Verify != nil {
		merged.Verify = override.Verify
	}
	if override.Proxy.URL != "" || len(override.Proxy.URLs) > 0 || override.Proxy.Auth != nil {
		merged.Proxy = override.Proxy
	}
	if override.ProxyRotator != nil {
		merged.ProxyRotator = override.ProxyRotator
	}
	if override.StealthyHeaders != nil {
		merged.StealthyHeaders = override.StealthyHeaders
	}
	if override.Impersonate != "" {
		merged.Impersonate = override.Impersonate
	}
	if override.HTTP3 != nil {
		merged.HTTP3 = override.HTTP3
	}
	if len(override.CookieValues) > 0 {
		merged.CookieValues = override.CookieValues
	}
	if len(override.Cookies) > 0 {
		merged.Cookies = override.Cookies
	}
	if override.Store != nil {
		merged.Store = override.Store
	}
	if override.FollowRedirects != "" {
		merged.FollowRedirects = override.FollowRedirects
	}
	if override.MaxRedirects != 0 {
		merged.MaxRedirects = override.MaxRedirects
	}
	if override.Timeout > 0 {
		merged.Timeout = override.Timeout
	}
	if override.Retries > 0 {
		merged.Retries = override.Retries
	}
	if override.RetryDelay > 0 {
		merged.RetryDelay = override.RetryDelay
	}
	if override.Context != nil {
		merged.Context = override.Context
	}
	if override.Safety.ObeyRobots || override.Safety.RobotsUserAgent != "" || override.Safety.BlockPrivateNetworks || len(override.Safety.BlockedCIDRs) > 0 {
		merged.Safety = override.Safety
	}
	return merged
}

func applyStaticProxyMeta(opts *fetchers.RequestOptions, meta map[string]any) {
	if opts == nil || meta == nil {
		return
	}
	if proxy, ok := meta["proxy"].(string); ok && strings.TrimSpace(proxy) != "" {
		opts.Proxy = fetchers.ProxyOptions{URL: proxy}
		opts.ProxyRotator = nil
	}
	if proxies, ok := meta["proxies"].(map[string]string); ok && len(proxies) > 0 {
		opts.Proxy = fetchers.ProxyOptions{URLs: cloneStringMap(proxies)}
		opts.ProxyRotator = nil
	}
}

func mergeSpiderBrowserOptions(base, override browser.BrowserOptions) browser.BrowserOptions {
	merged := cloneSpiderBrowserOptions(base)
	merged.Headers = mergeHTTPHeaders(base.Headers, override.Headers)
	if override.UserAgent != "" {
		merged.UserAgent = override.UserAgent
	}
	if len(override.Cookies) > 0 {
		merged.Cookies = cloneHTTPCookies(override.Cookies)
	}
	if override.Locale != "" {
		merged.Locale = override.Locale
	}
	if override.TimezoneID != "" {
		merged.TimezoneID = override.TimezoneID
	}
	if override.Proxy.Server != "" || override.Proxy.Username != "" || override.Proxy.Password != "" {
		merged.Proxy = override.Proxy
	}
	if override.CDPURL != "" {
		merged.CDPURL = override.CDPURL
	}
	if override.RealChrome {
		merged.RealChrome = true
	}
	if override.Headless {
		merged.Headless = true
	}
	if override.DisableResources {
		merged.DisableResources = true
	}
	if len(override.BlockedDomains) > 0 {
		merged.BlockedDomains = append(append([]string(nil), merged.BlockedDomains...), override.BlockedDomains...)
	}
	if override.BlockAds {
		merged.BlockAds = true
	}
	if override.DNSOverHTTPS {
		merged.DNSOverHTTPS = true
	}
	if override.NetworkIdle {
		merged.NetworkIdle = true
	}
	if override.LoadDOM {
		merged.LoadDOM = true
	}
	if override.Timeout > 0 {
		merged.Timeout = override.Timeout
	}
	if override.Wait > 0 {
		merged.Wait = override.Wait
	}
	if override.WaitSelector.Selector != "" || override.WaitSelector.State != "" {
		merged.WaitSelector = override.WaitSelector
	}
	if len(override.Actions) > 0 {
		merged.Actions = append(append([]browser.BrowserAction(nil), merged.Actions...), override.Actions...)
	}
	if override.CaptureXHR != "" {
		merged.CaptureXHR = override.CaptureXHR
	}
	if override.Screenshot.Enabled || override.Screenshot.FullPage || override.Screenshot.Selector != "" || override.Screenshot.Quality != 0 {
		merged.Screenshot = override.Screenshot
	}
	if len(override.ExtraFlags) > 0 {
		merged.ExtraFlags = append(append([]string(nil), merged.ExtraFlags...), override.ExtraFlags...)
	}
	merged.Stealth = mergeSpiderStealthOptions(merged.Stealth, override.Stealth)
	if override.Store != nil {
		merged.Store = override.Store
	}
	return merged
}

func mergeSpiderStealthOptions(base, override browser.BrowserStealthOptions) browser.BrowserStealthOptions {
	merged := base
	if override.Enabled {
		merged.Enabled = true
	}
	if override.GenerateHeaders {
		merged.GenerateHeaders = true
	}
	if override.GoogleReferer {
		merged.GoogleReferer = true
	}
	if override.HideCanvas {
		merged.HideCanvas = true
	}
	if override.BlockWebRTC {
		merged.BlockWebRTC = true
	}
	if override.DisableWebGL {
		merged.DisableWebGL = true
	}
	if override.SolveCloudflare {
		merged.SolveCloudflare = true
	}
	return merged
}

func cloneSpiderBrowserOptions(opts browser.BrowserOptions) browser.BrowserOptions {
	opts.Headers = opts.Headers.Clone()
	opts.Cookies = cloneHTTPCookies(opts.Cookies)
	opts.BlockedDomains = append([]string(nil), opts.BlockedDomains...)
	opts.Actions = append([]browser.BrowserAction(nil), opts.Actions...)
	opts.ExtraFlags = append([]string(nil), opts.ExtraFlags...)
	return opts
}

func browserProxyFromMeta(meta map[string]any) (browser.BrowserProxyOptions, bool) {
	if meta == nil {
		return browser.BrowserProxyOptions{}, false
	}
	if proxy, ok := meta["proxy"].(string); ok && strings.TrimSpace(proxy) != "" {
		return browser.BrowserProxyOptions{Server: proxy}, true
	}
	if values, ok := meta["proxy"].(map[string]string); ok {
		return browserProxyFromStringMap(values)
	}
	return browser.BrowserProxyOptions{}, false
}

func browserProxyFromStringMap(values map[string]string) (browser.BrowserProxyOptions, bool) {
	if len(values) == 0 {
		return browser.BrowserProxyOptions{}, false
	}
	lookup := make(map[string]string, len(values))
	for key, value := range values {
		lookup[strings.ToLower(strings.TrimSpace(key))] = value
	}
	server := strings.TrimSpace(lookup["server"])
	if server == "" {
		return browser.BrowserProxyOptions{}, false
	}
	return browser.BrowserProxyOptions{Server: server, Username: lookup["username"], Password: lookup["password"]}, true
}

func browserProxyFromFetchProxy(proxy fetchers.ProxyOptions) browser.BrowserProxyOptions {
	out := browser.BrowserProxyOptions{Server: proxy.URL}
	if out.Server == "" {
		if proxy.URLs["https"] != "" {
			out.Server = proxy.URLs["https"]
		} else if proxy.URLs["http"] != "" {
			out.Server = proxy.URLs["http"]
		}
	}
	if proxy.Auth != nil {
		out.Username = proxy.Auth.Username
		out.Password = proxy.Auth.Password
	}
	return out
}

func responseWithMergedMeta(response *goscrapling.Response, extra map[string]any) (*goscrapling.Response, error) {
	if response == nil || len(extra) == 0 {
		return response, nil
	}
	meta := response.Meta()
	if meta == nil {
		meta = map[string]any{}
	}
	for key, value := range extra {
		meta[key] = value
	}
	return goscrapling.NewResponse(bytes.NewReader(response.Body()), goscrapling.ResponseOptions{
		URL:         response.URL(),
		StatusCode:  response.StatusCode(),
		Reason:      response.Reason(),
		Headers:     response.Headers(),
		Request:     response.Request(),
		Encoding:    response.Encoding(),
		Cookies:     response.Cookies(),
		History:     response.History(),
		Meta:        meta,
		CapturedXHR: response.CapturedXHR(),
	})
}

func mergeHTTPHeaders(base, override http.Header) http.Header {
	merged := base.Clone()
	if merged == nil {
		merged = http.Header{}
	}
	for key, values := range override {
		merged.Del(key)
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	return merged
}

func cloneHTTPCookies(cookies []*http.Cookie) []*http.Cookie {
	if len(cookies) == 0 {
		return nil
	}
	cloned := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		copy := *cookie
		cloned = append(cloned, &copy)
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
