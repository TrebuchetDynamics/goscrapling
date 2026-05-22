package browser

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

var ErrBrowserSessionClosed = errors.New("browser session closed")

type browserPageState string

const (
	browserPageReady browserPageState = "ready"
	browserPageBusy  browserPageState = "busy"
	browserPageError browserPageState = "error"
)

type BrowserSessionOptions struct {
	Engine   BrowserEngine
	MaxPages int
	Options  BrowserOptions
}

type BrowserSessionStats struct {
	TotalPages int
	BusyPages  int
	FreePages  int
	ErrorPages int
	MaxPages   int
	Closed     bool
}

type BrowserSession struct {
	engine    BrowserEngine
	options   BrowserOptions
	maxPages  int
	available chan *browserSessionPage
	done      chan struct{}

	mu     sync.Mutex
	pages  []*browserSessionPage
	closed bool
}

type browserSessionPage struct {
	state browserPageState
}

func NewBrowserSession(opts BrowserSessionOptions) (*BrowserSession, error) {
	if opts.Engine == nil {
		return nil, ErrMissingBrowserEngine
	}
	maxPages := opts.MaxPages
	if maxPages <= 0 {
		maxPages = 1
	}
	return &BrowserSession{
		engine:    opts.Engine,
		options:   cloneBrowserOptions(opts.Options),
		maxPages:  maxPages,
		available: make(chan *browserSessionPage, maxPages),
		done:      make(chan struct{}),
	}, nil
}

func (s *BrowserSession) Fetch(ctx context.Context, rawURL string, opts BrowserOptions) (*Response, error) {
	if s == nil {
		return nil, ErrBrowserSessionClosed
	}
	if ctx == nil {
		ctx = context.Background()
	}
	page, err := s.acquirePage(ctx)
	if err != nil {
		return nil, err
	}

	merged := mergeBrowserOptions(s.options, opts)
	response, err := (BrowserFetcher{Engine: s.engine}).Fetch(ctx, rawURL, merged)
	if err != nil {
		s.releasePage(page, browserPageError)
		return nil, err
	}
	s.releasePage(page, browserPageReady)
	return response, nil
}

func (s *BrowserSession) Stats() BrowserSessionStats {
	if s == nil {
		return BrowserSessionStats{Closed: true}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := BrowserSessionStats{TotalPages: len(s.pages), MaxPages: s.maxPages, Closed: s.closed}
	for _, page := range s.pages {
		switch page.state {
		case browserPageBusy:
			stats.BusyPages++
		case browserPageError:
			stats.ErrorPages++
		default:
			stats.FreePages++
		}
	}
	return stats
}

func (s *BrowserSession) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	s.pages = nil
	close(s.done)
	return nil
}

func (s *BrowserSession) acquirePage(ctx context.Context) (*browserSessionPage, error) {
	for {
		select {
		case page := <-s.available:
			if err := s.markPageBusy(page); err != nil {
				return nil, err
			}
			return page, nil
		default:
		}

		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return nil, ErrBrowserSessionClosed
		}
		s.cleanupErrorPagesLocked()
		if len(s.pages) < s.maxPages {
			page := &browserSessionPage{state: browserPageBusy}
			s.pages = append(s.pages, page)
			s.mu.Unlock()
			return page, nil
		}
		s.mu.Unlock()

		select {
		case page := <-s.available:
			if err := s.markPageBusy(page); err != nil {
				return nil, err
			}
			return page, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.done:
			return nil, ErrBrowserSessionClosed
		}
	}
}

func (s *BrowserSession) markPageBusy(page *browserSessionPage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrBrowserSessionClosed
	}
	if page.state != browserPageReady {
		return ErrBrowserSessionClosed
	}
	page.state = browserPageBusy
	return nil
}

func (s *BrowserSession) releasePage(page *browserSessionPage, state browserPageState) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	page.state = state
	s.mu.Unlock()

	if state == browserPageReady {
		select {
		case s.available <- page:
		case <-s.done:
		}
	}
}

func (s *BrowserSession) cleanupErrorPagesLocked() {
	kept := s.pages[:0]
	for _, page := range s.pages {
		if page.state != browserPageError {
			kept = append(kept, page)
		}
	}
	s.pages = kept
}

func mergeBrowserOptions(defaults, overrides BrowserOptions) BrowserOptions {
	merged := cloneBrowserOptions(defaults)
	merged.Headers = mergeHeaders(defaults.Headers, overrides.Headers)
	if overrides.Headless {
		merged.Headless = true
	}
	if overrides.DisableResources {
		merged.DisableResources = true
	}
	if len(overrides.BlockedDomains) > 0 {
		merged.BlockedDomains = append(merged.BlockedDomains, overrides.BlockedDomains...)
	}
	if overrides.NetworkIdle {
		merged.NetworkIdle = true
	}
	if overrides.LoadDOM {
		merged.LoadDOM = true
	}
	if overrides.Timeout > 0 {
		merged.Timeout = overrides.Timeout
	}
	if overrides.Wait > 0 {
		merged.Wait = overrides.Wait
	}
	if overrides.WaitSelector.Selector != "" || overrides.WaitSelector.State != "" {
		merged.WaitSelector = overrides.WaitSelector
	}
	if len(overrides.Actions) > 0 {
		merged.Actions = append(merged.Actions, overrides.Actions...)
	}
	if overrides.Store != nil {
		merged.Store = overrides.Store
	}
	return merged
}

func cloneBrowserOptions(opts BrowserOptions) BrowserOptions {
	cloned := opts
	cloned.Headers = opts.Headers.Clone()
	cloned.BlockedDomains = append([]string(nil), opts.BlockedDomains...)
	cloned.Actions = append([]BrowserAction(nil), opts.Actions...)
	return cloned
}

func mergeHeaders(defaults, overrides http.Header) http.Header {
	merged := defaults.Clone()
	if merged == nil {
		merged = http.Header{}
	}
	for key, values := range overrides {
		merged.Del(key)
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	return merged
}
