package goscrapling

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ProxyOptions struct {
	URL  string
	URLs map[string]string
	Auth *BasicAuth
}

type proxyConfig struct {
	all      *proxyTarget
	byScheme map[string]proxyTarget
}

type proxyTarget struct {
	raw string
	url *url.URL
}

type proxyTracker struct {
	selected string
}

func (opts ProxyOptions) hasValues() bool {
	return opts.URL != "" || len(opts.URLs) > 0 || opts.Auth != nil
}

func newProxyConfig(opts ProxyOptions) (proxyConfig, bool, error) {
	if !opts.hasValues() {
		return proxyConfig{}, false, nil
	}
	if opts.URL != "" && len(opts.URLs) > 0 {
		return proxyConfig{}, false, fmt.Errorf("%w: Proxy URL and scheme-specific proxy URLs are mutually exclusive", ErrRequestOptions)
	}
	if opts.URL == "" && len(opts.URLs) == 0 {
		return proxyConfig{}, false, fmt.Errorf("%w: proxy auth requires a proxy URL", ErrRequestOptions)
	}

	if opts.URL != "" {
		target, err := parseProxyTarget(opts.URL, opts.Auth)
		if err != nil {
			return proxyConfig{}, false, err
		}
		return proxyConfig{all: &target}, true, nil
	}

	byScheme := make(map[string]proxyTarget, len(opts.URLs))
	for scheme, rawURL := range opts.URLs {
		scheme = strings.ToLower(strings.TrimSpace(scheme))
		if scheme != "http" && scheme != "https" {
			return proxyConfig{}, false, fmt.Errorf("%w: unsupported proxy map scheme %q", ErrRequestOptions, scheme)
		}
		target, err := parseProxyTarget(rawURL, opts.Auth)
		if err != nil {
			return proxyConfig{}, false, err
		}
		byScheme[scheme] = target
	}
	return proxyConfig{byScheme: byScheme}, true, nil
}

func parseProxyTarget(rawURL string, auth *BasicAuth) (proxyTarget, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return proxyTarget{}, fmt.Errorf("%w: parse proxy URL: %w", ErrRequestOptions, err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return proxyTarget{}, fmt.Errorf("%w: proxy URL must be absolute http(s)", ErrRequestOptions)
	}

	cloned := *parsed
	if auth != nil {
		cloned.User = url.UserPassword(auth.Username, auth.Password)
	}
	return proxyTarget{raw: rawURL, url: &cloned}, nil
}

func (c proxyConfig) proxyFunc(tracker *proxyTracker) func(*http.Request) (*url.URL, error) {
	return func(request *http.Request) (*url.URL, error) {
		if request == nil || request.URL == nil {
			return nil, nil
		}

		target, ok := c.targetForScheme(request.URL.Scheme)
		if !ok {
			return nil, nil
		}
		if tracker != nil {
			tracker.selected = target.raw
		}
		cloned := *target.url
		return &cloned, nil
	}
}

func (c proxyConfig) targetForScheme(scheme string) (proxyTarget, bool) {
	if c.all != nil {
		return *c.all, true
	}
	target, ok := c.byScheme[strings.ToLower(scheme)]
	return target, ok
}

func (t *proxyTracker) Selected() string {
	if t == nil {
		return ""
	}
	return t.selected
}

func cloneProxyOptions(opts ProxyOptions) ProxyOptions {
	cloned := ProxyOptions{
		URL: opts.URL,
	}
	if opts.Auth != nil {
		auth := *opts.Auth
		cloned.Auth = &auth
	}
	if len(opts.URLs) > 0 {
		cloned.URLs = make(map[string]string, len(opts.URLs))
		for scheme, proxyURL := range opts.URLs {
			cloned.URLs[scheme] = proxyURL
		}
	}
	return cloned
}
