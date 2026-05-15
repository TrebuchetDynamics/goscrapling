package spiders

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/goscrapling"
)

func TestSpiderEngineConcurrency(t *testing.T) {
	t.Run("limits global active fetches and applies backpressure", func(t *testing.T) {
		session := newControlledSession()
		sessions := NewSessionManager()
		if err := sessions.Add("default", session, SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}

		crawler := Crawler{
			Sessions:           sessions,
			ConcurrentRequests: 2,
		}
		done := runCrawler(t, crawler, []Request{
			{URL: "https://a.example/1"},
			{URL: "https://b.example/1"},
			{URL: "https://c.example/1"},
			{URL: "https://d.example/1"},
		})

		first := session.waitStarted(t)
		second := session.waitStarted(t)
		session.assertNoStart(t)

		session.release(first.URL)
		session.release(second.URL)
		third := session.waitStarted(t)
		fourth := session.waitStarted(t)
		session.release(third.URL)
		session.release(fourth.URL)

		result := waitCrawler(t, done)
		if result.err != nil {
			t.Fatalf("Run returned error: %v", result.err)
		}
		if result.result.Stats.Requests != 4 {
			t.Fatalf("requests = %d, want 4", result.result.Stats.Requests)
		}
		if session.maxActive() != 2 {
			t.Fatalf("max active fetches = %d, want 2", session.maxActive())
		}
	})

	t.Run("limits active fetches per domain", func(t *testing.T) {
		session := newControlledSession()
		sessions := NewSessionManager()
		if err := sessions.Add("default", session, SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}

		crawler := Crawler{
			Sessions:                    sessions,
			ConcurrentRequests:          3,
			ConcurrentRequestsPerDomain: 1,
		}
		done := runCrawler(t, crawler, []Request{
			{URL: "https://a.example/1"},
			{URL: "https://a.example/2"},
			{URL: "https://b.example/1"},
		})

		first := session.waitStarted(t)
		second := session.waitStarted(t)
		startedByDomain := map[string]controlledStart{
			testDomain(first.URL):  first,
			testDomain(second.URL): second,
		}
		if _, ok := startedByDomain["a.example"]; !ok {
			t.Fatalf("first two starts = %#v and %#v, want one a.example request", first, second)
		}
		if _, ok := startedByDomain["b.example"]; !ok {
			t.Fatalf("first two starts = %#v and %#v, want one b.example request", first, second)
		}
		session.assertNoStart(t)

		startedA := startedByDomain["a.example"].URL
		wantThird := "https://a.example/2"
		if startedA == wantThird {
			wantThird = "https://a.example/1"
		}

		session.release(startedA)
		third := session.waitStarted(t)
		if third.URL != wantThird {
			t.Fatalf("third start = %q, want serialized second a.example request", third.URL)
		}

		session.release(startedByDomain["b.example"].URL)
		session.release(third.URL)
		result := waitCrawler(t, done)
		if result.err != nil {
			t.Fatalf("Run returned error: %v", result.err)
		}
		if result.result.Stats.Requests != 3 {
			t.Fatalf("requests = %d, want 3", result.result.Stats.Requests)
		}
		if got := session.maxActiveForDomain("a.example"); got != 1 {
			t.Fatalf("max active for a.example = %d, want 1", got)
		}
	})

	t.Run("applies deterministic download delay and cancels active work", func(t *testing.T) {
		session := newControlledSession()
		sessions := NewSessionManager()
		if err := sessions.Add("default", session, SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		delayCalls := make(chan time.Duration, 3)
		crawler := Crawler{
			Sessions:           sessions,
			ConcurrentRequests: 2,
			DownloadDelay:      25 * time.Millisecond,
			sleep: func(ctx context.Context, delay time.Duration) error {
				delayCalls <- delay
				<-ctx.Done()
				return ctx.Err()
			},
		}
		done := runCrawlerWithContext(t, ctx, crawler, []Request{
			{URL: "https://a.example/1"},
			{URL: "https://b.example/1"},
			{URL: "https://c.example/1"},
		})

		if got := waitDelay(t, delayCalls); got != 25*time.Millisecond {
			t.Fatalf("delay = %s, want 25ms", got)
		}
		if got := waitDelay(t, delayCalls); got != 25*time.Millisecond {
			t.Fatalf("delay = %s, want 25ms", got)
		}
		assertNoDelay(t, delayCalls)
		session.assertNoStart(t)

		cancel()
		result := waitCrawler(t, done)
		if !errors.Is(result.err, context.Canceled) {
			t.Fatalf("Run error = %v, want context.Canceled", result.err)
		}
		if result.result.Stats.Requests != 0 {
			t.Fatalf("requests = %d, want 0 because delay was canceled before fetch", result.result.Stats.Requests)
		}
	})
}

type crawlerRunResult struct {
	result Result
	err    error
}

type controlledStart struct {
	URL    string
	Domain string
}

type controlledSession struct {
	mu          sync.Mutex
	requests    []Request
	active      int
	max         int
	activeByDom map[string]int
	maxByDom    map[string]int
	releases    map[string]chan struct{}
	started     chan controlledStart
}

func newControlledSession() *controlledSession {
	return &controlledSession{
		activeByDom: make(map[string]int),
		maxByDom:    make(map[string]int),
		releases:    make(map[string]chan struct{}),
		started:     make(chan controlledStart, 16),
	}
}

func (s *controlledSession) Fetch(ctx context.Context, request Request) (*goscrapling.Response, error) {
	domain := testDomain(request.URL)
	release := make(chan struct{})

	s.mu.Lock()
	s.requests = append(s.requests, request.clone())
	s.active++
	if s.active > s.max {
		s.max = s.active
	}
	s.activeByDom[domain]++
	if s.activeByDom[domain] > s.maxByDom[domain] {
		s.maxByDom[domain] = s.activeByDom[domain]
	}
	s.releases[request.URL] = release
	s.mu.Unlock()

	s.started <- controlledStart{URL: request.URL, Domain: domain}

	select {
	case <-release:
	case <-ctx.Done():
		s.finish(request.URL, domain)
		return nil, ctx.Err()
	}

	s.finish(request.URL, domain)
	return goscrapling.NewResponse(stringsReader("ok"), goscrapling.ResponseOptions{
		URL:        request.URL,
		StatusCode: http.StatusOK,
		Request: goscrapling.RequestMetadata{
			Method: request.MethodOrDefault(),
			URL:    request.URL,
		},
	})
}

func (s *controlledSession) finish(rawURL, domain string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active--
	s.activeByDom[domain]--
	delete(s.releases, rawURL)
}

func (s *controlledSession) waitStarted(t *testing.T) controlledStart {
	t.Helper()
	select {
	case started := <-s.started:
		return started
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for request start")
		return controlledStart{}
	}
}

func (s *controlledSession) assertNoStart(t *testing.T) {
	t.Helper()
	select {
	case started := <-s.started:
		t.Fatalf("unexpected request start: %#v", started)
	case <-time.After(30 * time.Millisecond):
	}
}

func (s *controlledSession) release(rawURL string) {
	s.mu.Lock()
	release := s.releases[rawURL]
	s.mu.Unlock()
	if release != nil {
		close(release)
	}
}

func (s *controlledSession) maxActive() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max
}

func (s *controlledSession) maxActiveForDomain(domain string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.maxByDom[domain]
}

func runCrawler(t *testing.T, crawler Crawler, requests []Request) <-chan crawlerRunResult {
	t.Helper()
	return runCrawlerWithContext(t, context.Background(), crawler, requests)
}

func runCrawlerWithContext(t *testing.T, ctx context.Context, crawler Crawler, requests []Request) <-chan crawlerRunResult {
	t.Helper()
	done := make(chan crawlerRunResult, 1)
	go func() {
		result, err := crawler.Run(ctx, requests)
		done <- crawlerRunResult{result: result, err: err}
	}()
	return done
}

func waitCrawler(t *testing.T, done <-chan crawlerRunResult) crawlerRunResult {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for crawler")
		return crawlerRunResult{}
	}
}

func waitDelay(t *testing.T, calls <-chan time.Duration) time.Duration {
	t.Helper()
	select {
	case delay := <-calls:
		return delay
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for delay")
		return 0
	}
}

func assertNoDelay(t *testing.T, calls <-chan time.Duration) {
	t.Helper()
	select {
	case delay := <-calls:
		t.Fatalf("unexpected delay call: %s", delay)
	case <-time.After(30 * time.Millisecond):
	}
}

func testDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func stringsReader(value string) *strings.Reader {
	return strings.NewReader(value)
}
