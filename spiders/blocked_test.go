package spiders_test

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestSpiderBlockedRetry(t *testing.T) {
	t.Run("default blocked status retries with accounting and proxy cleanup", func(t *testing.T) {
		session := &blockedSequenceSession{statuses: []int{http.StatusForbidden, http.StatusOK}}
		sessions := spiders.NewSessionManager()
		if err := sessions.Add("default", session, spiders.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}

		crawler := spiders.Crawler{
			Sessions:           sessions,
			MaxBlockedRetries:  1,
			ConcurrentRequests: 1,
			DefaultCallback: func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
				return []spiders.Output{spiders.Item(map[string]any{"status": response.StatusCode()})}, nil
			},
		}
		result, err := crawler.Run(context.Background(), []spiders.Request{{
			URL:      "https://example.com/protected",
			Priority: 10,
			Meta: map[string]any{
				"proxy":   "http://proxy-one.example:8080",
				"proxies": map[string]string{"https": "http://proxy-one.example:8080"},
			},
		}})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		if result.Stats.Requests != 2 || result.Stats.BlockedRequests != 1 || result.Stats.BlockedRetries != 1 {
			t.Fatalf("stats = %#v", result.Stats)
		}
		if !reflect.DeepEqual(result.Items, []map[string]any{{"status": http.StatusOK}}) {
			t.Fatalf("items = %#v", result.Items)
		}
		if len(session.requests) != 2 {
			t.Fatalf("requests = %d, want 2", len(session.requests))
		}
		retry := session.requests[1]
		if retry.RetryCount != 1 || !retry.DontFilter || retry.Priority != 9 {
			t.Fatalf("retry request = %#v", retry)
		}
		if _, ok := retry.Meta["proxy"]; ok {
			t.Fatalf("retry proxy metadata was not cleared: %#v", retry.Meta)
		}
		if _, ok := retry.Meta["proxies"]; ok {
			t.Fatalf("retry proxies metadata was not cleared: %#v", retry.Meta)
		}
	})

	t.Run("custom blocked and retry hooks can reroute the request", func(t *testing.T) {
		session := &blockedSequenceSession{statuses: []int{http.StatusOK, http.StatusOK}, bodies: []string{"captcha wall", "open"}}
		sessions := spiders.NewSessionManager()
		if err := sessions.Add("default", session, spiders.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}

		crawler := spiders.Crawler{
			Sessions:           sessions,
			MaxBlockedRetries:  2,
			ConcurrentRequests: 1,
			IsBlocked: func(_ context.Context, response spiders.Response) (bool, error) {
				return strings.Contains(response.Text(), "captcha"), nil
			},
			RetryBlockedRequest: func(_ context.Context, request spiders.Request, _ spiders.Response) (spiders.Request, error) {
				request.Headers.Set("X-Retry-Path", "stealth")
				request.Meta["retry_hook"] = true
				return request, nil
			},
			DefaultCallback: func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
				return []spiders.Output{spiders.Item(map[string]any{"body": response.Text(), "retry_hook": response.Request.Meta["retry_hook"]})}, nil
			},
		}
		result, err := crawler.Run(context.Background(), []spiders.Request{{URL: "https://example.com/custom"}})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		if result.Stats.BlockedRequests != 1 || result.Stats.BlockedRetries != 1 {
			t.Fatalf("stats = %#v", result.Stats)
		}
		if got := session.requests[1].Headers.Get("X-Retry-Path"); got != "stealth" {
			t.Fatalf("retry header = %q", got)
		}
		if !reflect.DeepEqual(result.Items, []map[string]any{{"body": "open", "retry_hook": true}}) {
			t.Fatalf("items = %#v", result.Items)
		}
	})

	t.Run("max blocked retries caps scheduling", func(t *testing.T) {
		session := &blockedSequenceSession{statuses: []int{http.StatusTooManyRequests, http.StatusTooManyRequests}}
		sessions := spiders.NewSessionManager()
		if err := sessions.Add("default", session, spiders.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}
		result, err := (spiders.Crawler{Sessions: sessions, MaxBlockedRetries: 1, ConcurrentRequests: 1}).Run(context.Background(), []spiders.Request{{URL: "https://example.com/rate-limit"}})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		if len(session.requests) != 2 {
			t.Fatalf("requests = %d, want original plus one retry", len(session.requests))
		}
		if result.Stats.BlockedRequests != 2 || result.Stats.BlockedRetries != 1 || result.Stats.Items != 0 {
			t.Fatalf("stats = %#v", result.Stats)
		}
	})
}

type blockedSequenceSession struct {
	mu       sync.Mutex
	statuses []int
	bodies   []string
	requests []spiders.Request
}

func (s *blockedSequenceSession) Fetch(_ context.Context, request spiders.Request) (*goscrapling.Response, error) {
	s.mu.Lock()
	index := len(s.requests)
	s.requests = append(s.requests, request)
	s.mu.Unlock()

	status := http.StatusOK
	if index < len(s.statuses) {
		status = s.statuses[index]
	}
	body := http.StatusText(status)
	if index < len(s.bodies) {
		body = s.bodies[index]
	}
	return goscrapling.NewResponse(strings.NewReader(body), goscrapling.ResponseOptions{
		URL:        request.URL,
		StatusCode: status,
		Request: goscrapling.RequestMetadata{
			Method:  request.MethodOrDefault(),
			URL:     request.URL,
			Headers: request.Headers,
		},
	})
}
