package fetchers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrUnsupportedStaticImpersonation = errors.New("static browser impersonation unsupported")
	ErrUnsupportedHTTP3               = errors.New("static HTTP/3 unsupported")
)

const defaultStaticIdentityUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

func validateStaticIdentityOptions(opts RequestOptions) error {
	if strings.TrimSpace(opts.Impersonate) != "" {
		return fmt.Errorf("%w: %w: net/http cannot impersonate browser TLS fingerprints", ErrRequestOptions, ErrUnsupportedStaticImpersonation)
	}
	if opts.HTTP3 != nil && *opts.HTTP3 {
		return fmt.Errorf("%w: %w: net/http transport does not provide request-scoped HTTP/3", ErrRequestOptions, ErrUnsupportedHTTP3)
	}
	return nil
}

func applyStaticIdentityHeaders(headers http.Header, opts RequestOptions) {
	if headers == nil || opts.StealthyHeaders == nil || !*opts.StealthyHeaders {
		return
	}
	setHeaderDefault(headers, "User-Agent", defaultStaticIdentityUserAgent)
	setHeaderDefault(headers, "Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	setHeaderDefault(headers, "Accept-Language", "en-US,en;q=0.9")
	setHeaderDefault(headers, "Referer", "https://www.google.com/")
	setHeaderDefault(headers, "Upgrade-Insecure-Requests", "1")
}

func setHeaderDefault(headers http.Header, key, value string) {
	if headers.Get(key) != "" {
		return
	}
	headers.Set(key, value)
}
