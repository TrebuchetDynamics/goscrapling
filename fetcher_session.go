package goscrapling

import (
	"net/http"
	"net/http/cookiejar"
)

type FetcherSessionOptions struct {
	Headers http.Header
	Client  *http.Client
	Proxy   ProxyOptions
	Store   Store
}

type FetcherSession struct {
	fetcher Fetcher
	headers http.Header
	proxy   ProxyOptions
	store   Store
}

func NewFetcherSession(opts FetcherSessionOptions) (*FetcherSession, error) {
	client, err := sessionHTTPClient(opts.Client)
	if err != nil {
		return nil, err
	}

	return &FetcherSession{
		fetcher: Fetcher{Client: client},
		headers: opts.Headers.Clone(),
		proxy:   cloneProxyOptions(opts.Proxy),
		store:   opts.Store,
	}, nil
}

func (s *FetcherSession) Get(url string, opts RequestOptions) (*Response, error) {
	return s.fetcher.Get(url, s.mergeOptions(opts))
}

func (s *FetcherSession) Post(url string, opts RequestOptions) (*Response, error) {
	return s.fetcher.Post(url, s.mergeOptions(opts))
}

func (s *FetcherSession) Put(url string, opts RequestOptions) (*Response, error) {
	return s.fetcher.Put(url, s.mergeOptions(opts))
}

func (s *FetcherSession) Delete(url string, opts RequestOptions) (*Response, error) {
	return s.fetcher.Delete(url, s.mergeOptions(opts))
}

func (s *FetcherSession) mergeOptions(opts RequestOptions) RequestOptions {
	if s == nil {
		return opts
	}

	mergedHeaders := s.headers.Clone()
	for key, values := range opts.Headers {
		mergedHeaders.Del(key)
		for _, value := range values {
			mergedHeaders.Add(key, value)
		}
	}

	opts.Headers = mergedHeaders
	if !opts.Proxy.hasValues() {
		opts.Proxy = cloneProxyOptions(s.proxy)
	}
	if opts.Store == nil {
		opts.Store = s.store
	}
	return opts
}

func sessionHTTPClient(client *http.Client) (*http.Client, error) {
	if client == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}

		transport := http.DefaultTransport
		if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = defaultTransport.Clone()
		}
		return &http.Client{Jar: jar, Transport: transport}, nil
	}

	cloned := *client
	if cloned.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		cloned.Jar = jar
	}
	return &cloned, nil
}
