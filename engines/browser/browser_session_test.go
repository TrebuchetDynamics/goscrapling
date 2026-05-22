package browser

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBrowserSessionPool(t *testing.T) {
	t.Run("reuses page slots and session defaults", func(t *testing.T) {
		engine := newSessionPoolEngine(nil)
		session, err := NewBrowserSession(BrowserSessionOptions{
			Engine:   engine,
			MaxPages: 1,
			Options: BrowserOptions{
				Headers: http.Header{"X-Session": []string{"default"}},
				LoadDOM: true,
			},
		})
		if err != nil {
			t.Fatalf("NewBrowserSession returned error: %v", err)
		}
		t.Cleanup(func() { _ = session.Close() })

		for _, path := range []string{"/one", "/two"} {
			response, err := session.Fetch(context.Background(), "https://example.com"+path, BrowserOptions{
				Headers: http.Header{"X-Request": []string{strings.TrimPrefix(path, "/")}},
			})
			if err != nil {
				t.Fatalf("Fetch(%s) returned error: %v", path, err)
			}
			first, ok := response.CSS(".page").First()
			if !ok {
				t.Fatalf("Fetch(%s) did not parse rendered page", path)
			}
			if got, ok := first.Attr("data-url"); !ok || got != "https://example.com"+path {
				t.Fatalf("Fetch(%s) data-url = %q ok=%v", path, got, ok)
			}
		}

		requireSessionStats(t, session.Stats(), BrowserSessionStats{
			TotalPages: 1,
			FreePages:  1,
			MaxPages:   1,
		})
		requests := engine.Requests()
		if len(requests) != 2 {
			t.Fatalf("engine saw %d requests, want 2", len(requests))
		}
		for i, request := range requests {
			if !request.LoadDOM {
				t.Fatalf("request %d did not inherit LoadDOM", i)
			}
			if got := request.Headers.Get("X-Session"); got != "default" {
				t.Fatalf("request %d X-Session = %q", i, got)
			}
			if got := request.Headers.Get("X-Request"); got == "" {
				t.Fatalf("request %d missing per-request X-Request header", i)
			}
		}
	})

	t.Run("bounds busy pages and queues until a page is free", func(t *testing.T) {
		release := make(chan struct{})
		engine := newSessionPoolEngine(release)
		session, err := NewBrowserSession(BrowserSessionOptions{Engine: engine, MaxPages: 2})
		if err != nil {
			t.Fatalf("NewBrowserSession returned error: %v", err)
		}
		t.Cleanup(func() { _ = session.Close() })

		errs := make(chan error, 3)
		for _, path := range []string{"/a", "/b"} {
			path := path
			go func() {
				_, err := session.Fetch(context.Background(), "https://example.com"+path, BrowserOptions{})
				errs <- err
			}()
		}
		waitStarted(t, engine.Started())
		waitStarted(t, engine.Started())
		requireSessionStats(t, session.Stats(), BrowserSessionStats{
			TotalPages: 2,
			BusyPages:  2,
			MaxPages:   2,
		})

		go func() {
			_, err := session.Fetch(context.Background(), "https://example.com/c", BrowserOptions{})
			errs <- err
		}()
		select {
		case url := <-engine.Started():
			t.Fatalf("queued request started before a page was free: %s", url)
		case <-time.After(40 * time.Millisecond):
		}

		close(release)
		for i := 0; i < 3; i++ {
			select {
			case err := <-errs:
				if err != nil {
					t.Fatalf("fetch %d returned error: %v", i, err)
				}
			case <-time.After(time.Second):
				t.Fatalf("fetch %d did not finish", i)
			}
		}
		if got := engine.MaxActive(); got != 2 {
			t.Fatalf("max active engine fetches = %d, want 2", got)
		}
		requireSessionStats(t, session.Stats(), BrowserSessionStats{
			TotalPages: 2,
			FreePages:  2,
			MaxPages:   2,
		})
	})

	t.Run("records error pages and closes cleanly", func(t *testing.T) {
		engine := newSessionPoolEngine(nil)
		engine.failSubstring = "/fail"
		session, err := NewBrowserSession(BrowserSessionOptions{Engine: engine, MaxPages: 2})
		if err != nil {
			t.Fatalf("NewBrowserSession returned error: %v", err)
		}

		_, err = session.Fetch(context.Background(), "https://example.com/fail", BrowserOptions{})
		if err == nil || !strings.Contains(err.Error(), "browser failure") {
			t.Fatalf("Fetch error = %v, want browser failure", err)
		}
		requireSessionStats(t, session.Stats(), BrowserSessionStats{
			TotalPages: 1,
			ErrorPages: 1,
			MaxPages:   2,
		})

		if err := session.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
		requireSessionStats(t, session.Stats(), BrowserSessionStats{
			MaxPages: 2,
			Closed:   true,
		})
		_, err = session.Fetch(context.Background(), "https://example.com/closed", BrowserOptions{})
		if !errors.Is(err, ErrBrowserSessionClosed) {
			t.Fatalf("Fetch after Close error = %v, want ErrBrowserSessionClosed", err)
		}
		if err := session.Close(); err != nil {
			t.Fatalf("second Close returned error: %v", err)
		}
	})
}

func requireSessionStats(t *testing.T, got, want BrowserSessionStats) {
	t.Helper()
	if got.TotalPages != want.TotalPages || got.BusyPages != want.BusyPages || got.FreePages != want.FreePages || got.ErrorPages != want.ErrorPages || got.MaxPages != want.MaxPages || got.Closed != want.Closed {
		t.Fatalf("session stats = %#v, want %#v", got, want)
	}
}

type sessionPoolEngine struct {
	mu            sync.Mutex
	started       chan string
	release       <-chan struct{}
	requests      []BrowserRequest
	active        int
	maxActive     int
	failSubstring string
}

func newSessionPoolEngine(release <-chan struct{}) *sessionPoolEngine {
	return &sessionPoolEngine{started: make(chan string, 16), release: release}
}

func (e *sessionPoolEngine) Started() <-chan string {
	return e.started
}

func (e *sessionPoolEngine) Requests() []BrowserRequest {
	e.mu.Lock()
	defer e.mu.Unlock()
	requests := make([]BrowserRequest, len(e.requests))
	copy(requests, e.requests)
	return requests
}

func (e *sessionPoolEngine) MaxActive() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.maxActive
}

func (e *sessionPoolEngine) Fetch(ctx context.Context, request BrowserRequest) (BrowserResult, error) {
	e.mu.Lock()
	e.requests = append(e.requests, request)
	e.active++
	if e.active > e.maxActive {
		e.maxActive = e.active
	}
	e.mu.Unlock()
	e.started <- request.URL
	defer func() {
		e.mu.Lock()
		e.active--
		e.mu.Unlock()
	}()

	if e.release != nil {
		select {
		case <-e.release:
		case <-ctx.Done():
			return BrowserResult{}, ctx.Err()
		}
	}
	if e.failSubstring != "" && strings.Contains(request.URL, e.failSubstring) {
		return BrowserResult{}, fmt.Errorf("browser failure for %s", request.URL)
	}
	return BrowserResult{
		URL:        request.URL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Body:       []byte(fmt.Sprintf(`<html><body><article class="page" data-url="%s">rendered</article></body></html>`, request.URL)),
	}, nil
}

func waitStarted(t *testing.T, started <-chan string) string {
	t.Helper()
	select {
	case url := <-started:
		return url
	case <-time.After(time.Second):
		t.Fatal("browser engine fetch did not start")
		return ""
	}
}
