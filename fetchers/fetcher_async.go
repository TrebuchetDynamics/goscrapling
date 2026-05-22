package fetchers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type ConcurrentFetcherOptions struct {
	Fetcher        Fetcher
	Session        *FetcherSession
	MaxConcurrency int
}

type ConcurrentFetcher struct {
	fetcher        Fetcher
	session        *FetcherSession
	maxConcurrency int
}

type ConcurrentRequest struct {
	Method  string
	URL     string
	Options RequestOptions
}

type ConcurrentResult struct {
	Request  ConcurrentRequest
	Response *Response
	Err      error
}

func NewConcurrentFetcher(opts ConcurrentFetcherOptions) *ConcurrentFetcher {
	return &ConcurrentFetcher{
		fetcher:        opts.Fetcher,
		session:        opts.Session,
		maxConcurrency: opts.MaxConcurrency,
	}
}

func (f *ConcurrentFetcher) Fetch(ctx context.Context, requests []ConcurrentRequest) []ConcurrentResult {
	results := make([]ConcurrentResult, len(requests))
	for i, request := range requests {
		results[i].Request = request
	}
	if len(requests) == 0 {
		return results
	}
	if ctx == nil {
		ctx = context.Background()
	}

	maxConcurrency := f.normalizedConcurrency(len(requests))
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(maxConcurrency)
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			defer wg.Done()
			for index := range jobs {
				if err := ctx.Err(); err != nil {
					results[index].Err = err
					continue
				}
				response, err := f.fetchOne(ctx, requests[index])
				results[index].Response = response
				results[index].Err = err
			}
		}()
	}

	for i := range requests {
		select {
		case <-ctx.Done():
			for j := i; j < len(results); j++ {
				results[j].Err = ctx.Err()
			}
			close(jobs)
			wg.Wait()
			return results
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()
	return results
}

func (f *ConcurrentFetcher) normalizedConcurrency(requestCount int) int {
	if f == nil || f.maxConcurrency <= 0 || f.maxConcurrency > requestCount {
		return requestCount
	}
	return f.maxConcurrency
}

func (f *ConcurrentFetcher) fetchOne(ctx context.Context, request ConcurrentRequest) (*Response, error) {
	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if method == "" {
		method = http.MethodGet
	}
	opts := request.Options
	opts.Context = ctx

	if f != nil && f.session != nil {
		switch method {
		case http.MethodGet:
			return f.session.Get(request.URL, opts)
		case http.MethodPost:
			return f.session.Post(request.URL, opts)
		case http.MethodPut:
			return f.session.Put(request.URL, opts)
		case http.MethodDelete:
			return f.session.Delete(request.URL, opts)
		default:
			return nil, fmt.Errorf("%w: unsupported concurrent fetch method %q", ErrRequestOptions, request.Method)
		}
	}

	fetcher := Fetcher{}
	if f != nil {
		fetcher = f.fetcher
	}
	switch method {
	case http.MethodGet:
		return fetcher.Get(request.URL, opts)
	case http.MethodPost:
		return fetcher.Post(request.URL, opts)
	case http.MethodPut:
		return fetcher.Put(request.URL, opts)
	case http.MethodDelete:
		return fetcher.Delete(request.URL, opts)
	default:
		return nil, fmt.Errorf("%w: unsupported concurrent fetch method %q", ErrRequestOptions, request.Method)
	}
}
