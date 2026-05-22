package browser

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var ErrBrowserOptions = errors.New("browser options")

type BrowserProxyOptions struct {
	Server   string
	Username string
	Password string
}

var supportedBrowserProxySchemes = map[string]struct{}{
	"http":   {},
	"https":  {},
	"socks4": {},
	"socks5": {},
}

var supportedBrowserCDPSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
	"ws":    {},
	"wss":   {},
}

var supportedBrowserWaitStates = map[BrowserWaitState]struct{}{
	"":                  {},
	BrowserWaitAttached: {},
	BrowserWaitDetached: {},
	BrowserWaitVisible:  {},
	BrowserWaitHidden:   {},
}

var supportedBrowserActionKinds = map[BrowserActionKind]struct{}{
	BrowserActionClick:           {},
	BrowserActionFill:            {},
	BrowserActionWaitForSelector: {},
	BrowserActionEvaluate:        {},
}

var browserDefaultBlockedResourcePatterns = []string{
	"*://*:*/*.avif",
	"*://*:*/*.css",
	"*://*:*/*.gif",
	"*://*:*/*.ico",
	"*://*:*/*.jpg",
	"*://*:*/*.jpeg",
	"*://*:*/*.mp3",
	"*://*:*/*.mp4",
	"*://*:*/*.png",
	"*://*:*/*.svg",
	"*://*:*/*.ttf",
	"*://*:*/*.webm",
	"*://*:*/*.webp",
	"*://*:*/*.woff",
	"*://*:*/*.woff2",
}

var browserOwnedAdDomains = []string{
	"2mdn.net",
	"doubleclick.net",
	"google-analytics.com",
	"googlesyndication.com",
	"googletagmanager.com",
}

func newBrowserRequest(rawURL string, opts BrowserOptions) (BrowserRequest, error) {
	if err := validateBrowserOptions(opts); err != nil {
		return BrowserRequest{}, err
	}
	headers := opts.Headers.Clone()
	if headers == nil {
		headers = http.Header{}
	}
	if opts.UserAgent != "" && headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", opts.UserAgent)
	}
	if opts.Locale != "" && headers.Get("Accept-Language") == "" {
		headers.Set("Accept-Language", opts.Locale)
	}

	return BrowserRequest{
		URL:              rawURL,
		Headers:          headers,
		UserAgent:        opts.UserAgent,
		Cookies:          cloneBrowserCookies(opts.Cookies),
		Locale:           opts.Locale,
		TimezoneID:       opts.TimezoneID,
		Proxy:            opts.Proxy,
		CDPURL:           opts.CDPURL,
		RealChrome:       opts.RealChrome,
		Headless:         opts.Headless,
		DisableResources: opts.DisableResources,
		BlockedDomains:   append([]string(nil), opts.BlockedDomains...),
		BlockAds:         opts.BlockAds,
		DNSOverHTTPS:     opts.DNSOverHTTPS,
		NetworkIdle:      opts.NetworkIdle,
		LoadDOM:          opts.LoadDOM,
		Timeout:          opts.Timeout,
		Wait:             opts.Wait,
		WaitSelector:     opts.WaitSelector,
		Actions:          append([]BrowserAction(nil), opts.Actions...),
		CaptureXHR:       opts.CaptureXHR,
		Screenshot:       opts.Screenshot,
		ExtraFlags:       append([]string(nil), opts.ExtraFlags...),
	}, nil
}

func validateBrowserOptions(opts BrowserOptions) error {
	if err := validateBrowserProxy(opts.Proxy); err != nil {
		return err
	}
	if opts.CDPURL != "" {
		parsed, err := url.Parse(opts.CDPURL)
		if err != nil || parsed.Host == "" {
			return fmt.Errorf("%w: invalid cdp_url %q", ErrBrowserOptions, opts.CDPURL)
		}
		if _, ok := supportedBrowserCDPSchemes[strings.ToLower(parsed.Scheme)]; !ok {
			return fmt.Errorf("%w: unsupported cdp_url scheme %q", ErrBrowserOptions, parsed.Scheme)
		}
	}
	for _, flag := range opts.ExtraFlags {
		if strings.TrimSpace(flag) == "" {
			return fmt.Errorf("%w: empty browser extra flag", ErrBrowserOptions)
		}
	}
	if _, ok := supportedBrowserWaitStates[opts.WaitSelector.State]; !ok {
		return fmt.Errorf("%w: unsupported wait selector state %q", ErrBrowserOptions, opts.WaitSelector.State)
	}
	if err := validateBrowserActions(opts.Actions); err != nil {
		return err
	}
	if opts.CaptureXHR != "" {
		if _, err := regexp.Compile(opts.CaptureXHR); err != nil {
			return fmt.Errorf("%w: invalid capture_xhr pattern: %v", ErrBrowserOptions, err)
		}
	}
	if err := validateBrowserScreenshot(opts.Screenshot); err != nil {
		return err
	}
	if strings.ContainsAny(opts.UserAgent, "\r\n") || strings.ContainsAny(opts.Locale, "\r\n") {
		return fmt.Errorf("%w: browser header values must not contain newlines", ErrBrowserOptions)
	}
	return nil
}

func validateBrowserActions(actions []BrowserAction) error {
	for _, action := range actions {
		if _, ok := supportedBrowserActionKinds[action.Kind]; !ok {
			return fmt.Errorf("%w: unsupported browser action %q", ErrBrowserOptions, action.Kind)
		}
		switch action.Kind {
		case BrowserActionClick, BrowserActionFill, BrowserActionWaitForSelector:
			if strings.TrimSpace(action.Selector) == "" {
				return fmt.Errorf("%w: browser action %q requires selector", ErrBrowserOptions, action.Kind)
			}
		case BrowserActionEvaluate:
			if strings.TrimSpace(action.Value) == "" {
				return fmt.Errorf("%w: browser evaluate action requires script", ErrBrowserOptions)
			}
		}
	}
	return nil
}

func validateBrowserScreenshot(screenshot BrowserScreenshotOptions) error {
	if !screenshot.Enabled {
		return nil
	}
	if screenshot.Quality < 0 || screenshot.Quality > 100 {
		return fmt.Errorf("%w: screenshot quality must be between 0 and 100", ErrBrowserOptions)
	}
	if strings.TrimSpace(screenshot.Selector) != "" && screenshot.FullPage {
		return fmt.Errorf("%w: screenshot selector and full_page are mutually exclusive", ErrBrowserOptions)
	}
	return nil
}

func validateBrowserProxy(proxy BrowserProxyOptions) error {
	if proxy.Server == "" {
		if proxy.Username != "" || proxy.Password != "" {
			return fmt.Errorf("%w: proxy credentials require a proxy server", ErrBrowserOptions)
		}
		return nil
	}
	parsed, err := url.Parse(proxy.Server)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("%w: invalid proxy server %q", ErrBrowserOptions, proxy.Server)
	}
	if _, ok := supportedBrowserProxySchemes[strings.ToLower(parsed.Scheme)]; !ok {
		return fmt.Errorf("%w: unsupported proxy scheme %q", ErrBrowserOptions, parsed.Scheme)
	}
	return nil
}

func cloneBrowserCookies(cookies []*http.Cookie) []*http.Cookie {
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

func browserBlockedURLPatterns(request BrowserRequest) []string {
	seen := map[string]struct{}{}
	var patterns []string
	add := func(pattern string) {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			return
		}
		if _, ok := seen[pattern]; ok {
			return
		}
		seen[pattern] = struct{}{}
		patterns = append(patterns, pattern)
	}
	if request.DisableResources {
		for _, pattern := range browserDefaultBlockedResourcePatterns {
			add(pattern)
		}
	}
	for _, domain := range request.BlockedDomains {
		for _, pattern := range browserDomainBlockPatterns(domain) {
			add(pattern)
		}
	}
	if request.BlockAds {
		for _, domain := range browserOwnedAdDomains {
			for _, pattern := range browserDomainBlockPatterns(domain) {
				add(pattern)
			}
		}
	}
	sort.Strings(patterns)
	return patterns
}

func browserDomainBlockPatterns(domain string) []string {
	domain = strings.Trim(strings.ToLower(strings.TrimSpace(domain)), ".")
	if domain == "" {
		return nil
	}
	return []string{"*://" + domain + "/*", "*://*." + domain + "/*"}
}
