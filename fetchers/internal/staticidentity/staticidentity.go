package staticidentity

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrUnsupportedStaticImpersonation = errors.New("static browser impersonation unsupported")
	ErrUnsupportedHTTP3               = errors.New("static HTTP/3 unsupported")
)

const DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

type Options struct {
	StealthyHeaders *bool
	Impersonate     string
	HTTP3           *bool
}

func ValidateOptions(opts Options) error {
	if strings.TrimSpace(opts.Impersonate) != "" {
		return ErrUnsupportedStaticImpersonation
	}
	if opts.HTTP3 != nil && *opts.HTTP3 {
		return ErrUnsupportedHTTP3
	}
	return nil
}

func ApplyHeaders(headers http.Header, opts Options) {
	if headers == nil || opts.StealthyHeaders == nil || !*opts.StealthyHeaders {
		return
	}
	setHeaderDefault(headers, "User-Agent", DefaultUserAgent)
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
