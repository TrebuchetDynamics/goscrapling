package fetchers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/TrebuchetDynamics/goscrapling/fetchers/internal/auth"
)

var ErrRequestOptions = errors.New("request options")

type BasicAuth = auth.BasicAuth

type RequestOptions struct {
	Headers         http.Header
	Body            io.Reader
	Params          url.Values
	Data            url.Values
	JSON            any
	Auth            *BasicAuth
	Verify          *bool
	Proxy           ProxyOptions
	ProxyRotator    *ProxyRotator
	StealthyHeaders *bool
	Impersonate     string
	HTTP3           *bool
	Safety          FetchSafetyOptions
	CookieValues    map[string]string
	Cookies         []*http.Cookie
	Context         context.Context

	Store           Store
	FollowRedirects RedirectPolicy
	MaxRedirects    int
	Timeout         time.Duration
	Retries         int
	RetryDelay      time.Duration
}

func Bool(value bool) *bool {
	return &value
}

func prepareRequest(rawURL string, opts RequestOptions) (string, []byte, http.Header, error) {
	if err := validateStaticIdentityOptions(opts); err != nil {
		return "", nil, nil, err
	}

	requestURL, err := appendRequestParams(rawURL, opts.Params)
	if err != nil {
		return "", nil, nil, err
	}

	headers := opts.Headers.Clone()
	if headers == nil {
		headers = http.Header{}
	}
	body, contentType, err := encodeRequestBody(opts)
	if err != nil {
		return "", nil, nil, err
	}
	if contentType != "" && headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", contentType)
	}
	applyStaticIdentityHeaders(headers, opts)

	return requestURL, body, headers, nil
}

func appendRequestParams(rawURL string, params url.Values) (string, error) {
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

func encodeRequestBody(opts RequestOptions) ([]byte, string, error) {
	bodyKinds := 0
	if opts.Body != nil {
		bodyKinds++
	}
	if len(opts.Data) > 0 {
		bodyKinds++
	}
	if opts.JSON != nil {
		bodyKinds++
	}
	if bodyKinds > 1 {
		return nil, "", fmt.Errorf("%w: Body, Data, and JSON are mutually exclusive", ErrRequestOptions)
	}

	switch {
	case opts.Body != nil:
		body, err := readRequestBody(opts.Body)
		if err != nil {
			return nil, "", err
		}
		return body, "", nil
	case len(opts.Data) > 0:
		return []byte(opts.Data.Encode()), "application/x-www-form-urlencoded", nil
	case opts.JSON != nil:
		body, err := json.Marshal(opts.JSON)
		if err != nil {
			return nil, "", fmt.Errorf("%w: marshal JSON: %w", ErrRequestOptions, err)
		}
		return body, "application/json", nil
	default:
		return nil, "", nil
	}
}

func applyRequestCookies(request *http.Request, opts RequestOptions) {
	if request == nil {
		return
	}

	explicitNames := map[string]struct{}{}
	for _, cookie := range opts.Cookies {
		if cookie == nil || cookie.Name == "" {
			continue
		}
		explicitNames[cookie.Name] = struct{}{}
	}

	names := make([]string, 0, len(opts.CookieValues))
	for name := range opts.CookieValues {
		if name == "" {
			continue
		}
		if _, explicit := explicitNames[name]; explicit {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		request.AddCookie(&http.Cookie{Name: name, Value: opts.CookieValues[name]})
	}

	for _, cookie := range opts.Cookies {
		if cookie == nil || cookie.Name == "" {
			continue
		}
		request.AddCookie(cookie)
	}
}
