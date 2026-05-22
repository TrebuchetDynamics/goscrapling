package fetchers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrentFetcher(t *testing.T) {
	t.Run("bounds concurrent fetches and preserves result order", func(t *testing.T) {
		var active atomic.Int64
		var maxActive atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current := active.Add(1)
			for {
				max := maxActive.Load()
				if current <= max || maxActive.CompareAndSwap(max, current) {
					break
				}
			}
			defer active.Add(-1)

			time.Sleep(40 * time.Millisecond)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<html><body><article class="result" data-path="%s">ok</article></body></html>`, r.URL.Path)
		}))
		t.Cleanup(server.Close)

		fetcher := NewConcurrentFetcher(ConcurrentFetcherOptions{MaxConcurrency: 2})
		results := fetcher.Fetch(context.Background(), []ConcurrentRequest{
			{URL: server.URL + "/one"},
			{URL: server.URL + "/two"},
			{URL: server.URL + "/three"},
		})

		if len(results) != 3 {
			t.Fatalf("results len = %d, want 3", len(results))
		}
		for i, wantPath := range []string{"/one", "/two", "/three"} {
			if results[i].Err != nil {
				t.Fatalf("result %d error = %v", i, results[i].Err)
			}
			first, ok := results[i].Response.CSS(".result").First()
			if !ok {
				t.Fatalf("result %d did not parse article", i)
			}
			if got, ok := first.Attr("data-path"); !ok || got != wantPath {
				t.Fatalf("result %d path = %q ok=%v, want %q", i, got, ok, wantPath)
			}
		}
		if got := maxActive.Load(); got != 2 {
			t.Fatalf("max concurrent requests = %d, want bounded overlap of 2", got)
		}
	})

	t.Run("collects per request errors beside successful responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="ok">ok</article></body></html>`)
		}))
		t.Cleanup(server.Close)

		fetcher := NewConcurrentFetcher(ConcurrentFetcherOptions{MaxConcurrency: 2})
		results := fetcher.Fetch(context.Background(), []ConcurrentRequest{
			{URL: server.URL + "/ok"},
			{URL: "://bad-url"},
		})

		if results[0].Err != nil || results[0].Response == nil || results[0].Response.CSS(".ok").Len() != 1 {
			t.Fatalf("successful result = %#v", results[0])
		}
		if results[1].Err == nil || results[1].Response != nil {
			t.Fatalf("error result = %#v, want err with nil response", results[1])
		}
	})

	t.Run("uses a shared session for default headers and cookies", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/seed":
				http.SetCookie(w, &http.Cookie{Name: "sid", Value: "abc123"})
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, `<html><body><article class="seed">seed</article></body></html>`)
			case "/echo-a", "/echo-b":
				cookie, err := r.Cookie("sid")
				if err != nil {
					t.Errorf("missing sid cookie: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(w, `<html><body><article class="echo" data-session="%s" data-cookie="%s"></article></body></html>`, r.Header.Get("X-Session"), cookie.Value)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(server.Close)

		session, err := NewFetcherSession(FetcherSessionOptions{
			Headers: http.Header{"X-Session": []string{"session-default"}},
		})
		if err != nil {
			t.Fatalf("NewFetcherSession: %v", err)
		}
		if _, err := session.Get(server.URL+"/seed", RequestOptions{Retries: 1}); err != nil {
			t.Fatalf("seed session cookie: %v", err)
		}

		fetcher := NewConcurrentFetcher(ConcurrentFetcherOptions{Session: session, MaxConcurrency: 2})
		results := fetcher.Fetch(context.Background(), []ConcurrentRequest{
			{URL: server.URL + "/echo-a"},
			{URL: server.URL + "/echo-b"},
		})

		for i, result := range results {
			if result.Err != nil {
				t.Fatalf("result %d error = %v", i, result.Err)
			}
			first, ok := result.Response.CSS(".echo").First()
			if !ok {
				t.Fatalf("result %d missing echo article", i)
			}
			if got, ok := first.Attr("data-session"); !ok || got != "session-default" {
				t.Fatalf("result %d session header = %q ok=%v", i, got, ok)
			}
			if got, ok := first.Attr("data-cookie"); !ok || got != "abc123" {
				t.Fatalf("result %d cookie = %q ok=%v", i, got, ok)
			}
		}
	})

	t.Run("cancels in flight requests and queued work", func(t *testing.T) {
		started := make(chan struct{})
		var queuedRequests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/blocked":
				close(started)
				<-r.Context().Done()
			case "/queued":
				queuedRequests.Add(1)
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(server.Close)

		ctx, cancel := context.WithCancel(context.Background())
		fetcher := NewConcurrentFetcher(ConcurrentFetcherOptions{MaxConcurrency: 1})
		resultCh := make(chan []ConcurrentResult, 1)
		go func() {
			resultCh <- fetcher.Fetch(ctx, []ConcurrentRequest{
				{URL: server.URL + "/blocked", Options: RequestOptions{Retries: 1}},
				{URL: server.URL + "/queued", Options: RequestOptions{Retries: 1}},
			})
		}()

		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("blocked request did not start")
		}
		cancel()

		var results []ConcurrentResult
		select {
		case results = <-resultCh:
		case <-time.After(time.Second):
			t.Fatal("concurrent fetch did not return after cancellation")
		}

		if len(results) != 2 {
			t.Fatalf("results len = %d, want 2", len(results))
		}
		if !errors.Is(results[0].Err, context.Canceled) {
			t.Fatalf("in-flight error = %v, want context.Canceled", results[0].Err)
		}
		if !errors.Is(results[1].Err, context.Canceled) {
			t.Fatalf("queued error = %v, want context.Canceled", results[1].Err)
		}
		if got := queuedRequests.Load(); got != 0 {
			t.Fatalf("queued endpoint saw %d requests, want 0", got)
		}
	})
}
