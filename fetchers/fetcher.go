package fetchers

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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
	requestURL, body, headers, err := prepareRequest(rawURL, opts)
	if err != nil {
		return nil, err
	}
	opts.Headers = headers

	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	if err := enforceFetchSafety(method, requestURL, opts, client); err != nil {
		return nil, err
	}

	attempts := retryAttempts(opts.Retries)
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		response, err := f.doAttempt(method, requestURL, body, opts)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if !isRetriableFetcherError(err) {
			return nil, err
		}

		if attempt < attempts && opts.RetryDelay > 0 {
			time.Sleep(opts.RetryDelay)
		}
	}

	if errors.Is(lastErr, ErrProxyRequest) {
		return nil, &FetcherError{
			Kind:     FetcherErrorProxy,
			Method:   method,
			URL:      requestURL,
			Attempts: attempts,
			Err:      lastErr,
		}
	}
	return nil, &FetcherError{
		Kind:     FetcherErrorRetryExhausted,
		Method:   method,
		URL:      requestURL,
		Attempts: attempts,
		Err:      ErrRetryExhausted,
	}
}

func (f Fetcher) doAttempt(method, rawURL string, body []byte, opts RequestOptions) (*Response, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
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
	applyRequestCookies(request, opts)
	if opts.Auth != nil && request.Header.Get("Authorization") == "" {
		request.SetBasicAuth(opts.Auth.Username, opts.Auth.Password)
	}

	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	var history []*Response
	proxy := &proxyTracker{}
	client, err = clientWithRedirectPolicy(client, opts, func(response *http.Response) {
		redirect, err := newResponseFromHTTPResponse(response, nil, opts.Store)
		if err == nil {
			history = append(history, redirect)
		}
	}, proxy)
	if err != nil {
		return nil, err
	}

	httpResponse, err := client.Do(request)
	if err != nil {
		if httpResponse != nil && httpResponse.Body != nil {
			httpResponse.Body.Close()
		}
		return nil, classifyRequestError(method, rawURL, err, proxy.Selected() != "")
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

	meta := map[string]any(nil)
	if selectedProxy := proxy.Selected(); selectedProxy != "" {
		meta = map[string]any{"proxy": selectedProxy}
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
		Cookies: httpResponse.Cookies(),
		History: history,
		Meta:    meta,
		Store:   opts.Store,
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

func clientWithRedirectPolicy(client *http.Client, opts RequestOptions, onRedirect func(*http.Response), proxy *proxyTracker) (*http.Client, error) {
	if opts.Proxy.hasValues() && opts.ProxyRotator != nil {
		return nil, fmt.Errorf("%w: Proxy and ProxyRotator are mutually exclusive", ErrRequestOptions)
	}

	cloned := *client
	if (opts.Verify != nil && !*opts.Verify) || opts.Proxy.hasValues() || opts.ProxyRotator != nil {
		transport, err := transportWithRequestOptions(cloned.Transport, opts, proxy)
		if err != nil {
			return nil, err
		}
		cloned.Transport = transport
	}
	cloned.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if onRedirect != nil && request != nil && request.Response != nil {
			onRedirect(request.Response)
		}

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
	return &cloned, nil
}

func transportWithRequestOptions(roundTripper http.RoundTripper, opts RequestOptions, tracker *proxyTracker) (http.RoundTripper, error) {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	transport, ok := roundTripper.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("%w: request transport options require an *http.Transport", ErrRequestOptions)
	}

	cloned := transport.Clone()
	if opts.Verify != nil && !*opts.Verify {
		if cloned.TLSClientConfig == nil {
			cloned.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			cloned.TLSClientConfig = cloned.TLSClientConfig.Clone()
			cloned.TLSClientConfig.InsecureSkipVerify = true
		}
	}
	proxyOptions := opts.Proxy
	if opts.ProxyRotator != nil {
		if opts.Proxy.hasValues() {
			return nil, fmt.Errorf("%w: Proxy and ProxyRotator are mutually exclusive", ErrRequestOptions)
		}
		rotatedProxy, err := opts.ProxyRotator.Next()
		if err != nil {
			return nil, err
		}
		proxyOptions = rotatedProxy
	}
	if proxyOptions.hasValues() {
		proxyConfig, _, err := newProxyConfig(proxyOptions)
		if err != nil {
			return nil, err
		}
		cloned.Proxy = proxyConfig.proxyFunc(tracker)
	}
	return cloned, nil
}

func newResponseFromHTTPResponse(httpResponse *http.Response, body []byte, store Store) (*Response, error) {
	if httpResponse == nil {
		return nil, nil
	}

	responseURL := ""
	request := RequestMetadata{}
	if httpResponse.Request != nil {
		request.Method = httpResponse.Request.Method
		request.Headers = httpResponse.Request.Header
		if httpResponse.Request.URL != nil {
			responseURL = httpResponse.Request.URL.String()
			request.URL = responseURL
		}
	}

	return NewResponse(bytes.NewReader(body), ResponseOptions{
		URL:        responseURL,
		StatusCode: httpResponse.StatusCode,
		Reason:     http.StatusText(httpResponse.StatusCode),
		Headers:    httpResponse.Header,
		Request:    request,
		Cookies:    httpResponse.Cookies(),
		Store:      store,
	})
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

func classifyRequestError(method, rawURL string, err error, proxyActive bool) error {
	switch {
	case proxyActive:
		return &FetcherError{
			Kind:   FetcherErrorProxy,
			Method: method,
			URL:    rawURL,
			Err:    fmt.Errorf("%w: %w", ErrProxyRequest, err),
		}
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
	case errors.Is(err, context.Canceled):
		return err
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
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, ErrPrivateAddressRedirect) ||
		errors.Is(err, ErrRedirectNotAllowed) ||
		errors.Is(err, ErrRequestOptions) ||
		errors.Is(err, ErrRequestTimeout) {
		return false
	}
	return true
}
