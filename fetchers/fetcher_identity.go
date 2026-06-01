package fetchers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TrebuchetDynamics/goscrapling/fetchers/internal/staticidentity"
)

var (
	ErrUnsupportedStaticImpersonation = staticidentity.ErrUnsupportedStaticImpersonation
	ErrUnsupportedHTTP3               = staticidentity.ErrUnsupportedHTTP3
)

const defaultStaticIdentityUserAgent = staticidentity.DefaultUserAgent

func validateStaticIdentityOptions(opts RequestOptions) error {
	err := staticidentity.ValidateOptions(staticidentity.Options{
		StealthyHeaders: opts.StealthyHeaders,
		Impersonate:     opts.Impersonate,
		HTTP3:           opts.HTTP3,
	})
	if errors.Is(err, ErrUnsupportedStaticImpersonation) {
		return fmt.Errorf("%w: %w: net/http cannot impersonate browser TLS fingerprints", ErrRequestOptions, err)
	}
	if errors.Is(err, ErrUnsupportedHTTP3) {
		return fmt.Errorf("%w: %w: net/http transport does not provide request-scoped HTTP/3", ErrRequestOptions, err)
	}
	return err
}

func applyStaticIdentityHeaders(headers http.Header, opts RequestOptions) {
	staticidentity.ApplyHeaders(headers, staticidentity.Options{
		StealthyHeaders: opts.StealthyHeaders,
		Impersonate:     opts.Impersonate,
		HTTP3:           opts.HTTP3,
	})
}

func setHeaderDefault(headers http.Header, key, value string) {
	if headers.Get(key) != "" {
		return
	}
	headers.Set(key, value)
}
