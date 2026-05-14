package goscrapling

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultRetryAttempts = 3

type RedirectPolicy string

const (
	RedirectPolicySafe RedirectPolicy = ""
	RedirectPolicyAll  RedirectPolicy = "all"
	RedirectPolicyNone RedirectPolicy = "none"
)

type Fetcher struct {
	Client *http.Client
}

type RequestOptions struct {
	Headers         http.Header
	Body            io.Reader
	Store           Store
	FollowRedirects RedirectPolicy
	MaxRedirects    int
	Timeout         time.Duration
	Retries         int
	RetryDelay      time.Duration
}

func (f Fetcher) Get(url string, opts RequestOptions) (*Response, error) {
	return f.do(http.MethodGet, url, opts)
}

func (f Fetcher) Post(url string, opts RequestOptions) (*Response, error) {
	return f.do(http.MethodPost, url, opts)
}

func (f Fetcher) Put(url string, opts RequestOptions) (*Response, error) {
	return f.do(http.MethodPut, url, opts)
}

func (f Fetcher) Delete(url string, opts RequestOptions) (*Response, error) {
	return f.do(http.MethodDelete, url, opts)
}

func (f Fetcher) do(method, rawURL string, opts RequestOptions) (*Response, error) {
	body, err := readRequestBody(opts.Body)
	if err != nil {
		return nil, err
	}

	attempts := retryAttempts(opts.Retries)
	for attempt := 1; attempt <= attempts; attempt++ {
		response, err := f.doAttempt(method, rawURL, body, opts)
		if err == nil {
			return response, nil
		}
		if !isRetriableFetcherError(err) {
			return nil, err
		}

		if attempt < attempts && opts.RetryDelay > 0 {
			time.Sleep(opts.RetryDelay)
		}
	}

	return nil, &FetcherError{
		Kind:     FetcherErrorRetryExhausted,
		Method:   method,
		URL:      rawURL,
		Attempts: attempts,
		Err:      ErrRetryExhausted,
	}
}

func (f Fetcher) doAttempt(method, rawURL string, body []byte, opts RequestOptions) (*Response, error) {
	ctx := context.Background()
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	request, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header = opts.Headers.Clone()

	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	client = clientWithRedirectPolicy(client, opts)

	httpResponse, err := client.Do(request)
	if err != nil {
		if httpResponse != nil && httpResponse.Body != nil {
			httpResponse.Body.Close()
		}
		return nil, classifyRequestError(method, rawURL, err)
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	responseURL := rawURL
	if httpResponse.Request != nil && httpResponse.Request.URL != nil {
		responseURL = httpResponse.Request.URL.String()
	}

	return NewResponse(bytes.NewReader(responseBody), ResponseOptions{
		URL:        responseURL,
		StatusCode: httpResponse.StatusCode,
		Reason:     http.StatusText(httpResponse.StatusCode),
		Headers:    httpResponse.Header,
		Request: RequestMetadata{
			Method:  method,
			URL:     rawURL,
			Headers: request.Header,
		},
		Store: opts.Store,
	})
}

func readRequestBody(body io.Reader) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return io.ReadAll(body)
}

func retryAttempts(retries int) int {
	if retries > 0 {
		return retries
	}
	return defaultRetryAttempts
}

func clientWithRedirectPolicy(client *http.Client, opts RequestOptions) *http.Client {
	cloned := *client
	cloned.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if opts.FollowRedirects == RedirectPolicyNone {
			return http.ErrUseLastResponse
		}

		maxRedirects := opts.MaxRedirects
		if maxRedirects == 0 {
			maxRedirects = 30
		}
		if maxRedirects > 0 && len(via) >= maxRedirects {
			return ErrRedirectNotAllowed
		}

		if opts.FollowRedirects == RedirectPolicyAll {
			return nil
		}
		if isPrivateAddressRedirect(request, via) {
			return ErrPrivateAddressRedirect
		}
		return nil
	}
	return &cloned
}

func isPrivateAddressRedirect(request *http.Request, via []*http.Request) bool {
	if request == nil || request.URL == nil || len(via) == 0 {
		return false
	}

	previous := via[len(via)-1]
	if previous != nil && previous.URL != nil && strings.EqualFold(request.URL.Hostname(), previous.URL.Hostname()) {
		return false
	}

	ip := net.ParseIP(request.URL.Hostname())
	return isPrivateIP(ip)
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func classifyRequestError(method, rawURL string, err error) error {
	switch {
	case errors.Is(err, ErrPrivateAddressRedirect):
		return &FetcherError{
			Kind:   FetcherErrorPrivateRedirect,
			Method: method,
			URL:    rawURL,
			Err:    ErrPrivateAddressRedirect,
		}
	case errors.Is(err, ErrRedirectNotAllowed):
		return &FetcherError{
			Kind:   FetcherErrorRedirect,
			Method: method,
			URL:    rawURL,
			Err:    ErrRedirectNotAllowed,
		}
	case errors.Is(err, context.DeadlineExceeded), os.IsTimeout(err):
		return &FetcherError{
			Kind:   FetcherErrorTimeout,
			Method: method,
			URL:    rawURL,
			Err:    ErrRequestTimeout,
		}
	default:
		return err
	}
}

func isRetriableFetcherError(err error) bool {
	if errors.Is(err, ErrPrivateAddressRedirect) ||
		errors.Is(err, ErrRedirectNotAllowed) ||
		errors.Is(err, ErrRequestTimeout) {
		return false
	}
	return true
}
