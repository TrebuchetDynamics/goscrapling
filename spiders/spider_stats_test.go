package spiders

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
)

func TestSpiderLifecycleStatsAndStream(t *testing.T) {
	session := &lifecycleSession{responses: map[string]lifecycleResponse{
		"https://example.com/list":   {status: http.StatusOK, body: "list-body"},
		"https://example.com/detail": {status: http.StatusCreated, body: "detail-body"},
	}}
	sessions := NewSessionManager()
	if err := sessions.Add("default", session, SessionOptions{Default: true}); err != nil {
		t.Fatalf("add session: %v", err)
	}

	var events []string
	crawler := Crawler{
		Sessions:           sessions,
		ConcurrentRequests: 1,
		OnStart: func(_ context.Context, resuming bool) error {
			events = append(events, "start", boolEvent(resuming))
			return nil
		},
		OnScrapedItem: func(_ context.Context, item map[string]any) (map[string]any, error) {
			events = append(events, "item:"+item["url"].(string))
			if item["drop"] == true {
				return nil, nil
			}
			item["hooked"] = true
			return item, nil
		},
		OnError: func(_ context.Context, request Request, err error) error {
			events = append(events, "error:"+request.URL+":"+err.Error())
			return nil
		},
		OnClose: func(_ context.Context, result Result) error {
			events = append(events, "close", boolEvent(result.Paused))
			return nil
		},
		DefaultCallback: func(_ context.Context, response Response) ([]Output, error) {
			switch response.URL() {
			case "https://example.com/list":
				return []Output{
					Item(map[string]any{"url": response.URL()}),
					Item(map[string]any{"url": "drop-me", "drop": true}),
					Next(Request{URL: "https://example.com/detail"}),
					Next(Request{URL: "https://example.com/fail"}),
				}, nil
			case "https://example.com/detail":
				return []Output{Item(map[string]any{"url": response.URL()})}, nil
			default:
				return nil, nil
			}
		},
	}

	items, done := crawler.Stream(context.Background(), []Request{{URL: "https://example.com/list"}})
	var streamed []map[string]any
	for item := range items {
		streamed = append(streamed, item)
	}
	streamResult := <-done
	if streamResult.Err != nil {
		t.Fatalf("Stream returned error: %v", streamResult.Err)
	}
	result := streamResult.Result

	wantItems := []map[string]any{
		{"url": "https://example.com/list", "hooked": true},
		{"url": "https://example.com/detail", "hooked": true},
	}
	if !reflect.DeepEqual(streamed, wantItems) {
		t.Fatalf("streamed items = %#v, want %#v", streamed, wantItems)
	}
	if !reflect.DeepEqual(result.Items, wantItems) {
		t.Fatalf("result items = %#v, want %#v", result.Items, wantItems)
	}
	wantEvents := []string{
		"start", "false",
		"item:https://example.com/list",
		"item:drop-me",
		"item:https://example.com/detail",
		"error:https://example.com/fail:boom",
		"close", "false",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}

	if result.Stats.Requests != 2 || result.Stats.Failed != 1 || result.Stats.Items != 2 || result.Stats.ItemsDropped != 1 {
		t.Fatalf("stats counts = %#v", result.Stats)
	}
	if got := result.Stats.Sessions["default"]; got != 2 {
		t.Fatalf("default session stats = %d, want 2", got)
	}
	if result.Stats.StatusCodes[http.StatusOK] != 1 || result.Stats.StatusCodes[http.StatusCreated] != 1 {
		t.Fatalf("status counters = %#v", result.Stats.StatusCodes)
	}
	if result.Stats.ResponseBytes != len("list-body")+len("detail-body") {
		t.Fatalf("response bytes = %d", result.Stats.ResponseBytes)
	}
	if result.Stats.DomainResponseBytes["example.com"] != result.Stats.ResponseBytes {
		t.Fatalf("domain response bytes = %#v", result.Stats.DomainResponseBytes)
	}
	if result.Stats.StartTime.IsZero() || result.Stats.EndTime.IsZero() || result.Stats.Elapsed < 0 {
		t.Fatalf("timing stats = start %s end %s elapsed %s", result.Stats.StartTime, result.Stats.EndTime, result.Stats.Elapsed)
	}

	dir := t.TempDir()
	if err := result.ItemList().ToJSON(filepath.Join(dir, "items.json"), true); err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}
	if err := result.ItemList().ToJSONL(filepath.Join(dir, "items.jsonl")); err != nil {
		t.Fatalf("ToJSONL returned error: %v", err)
	}
	jsonBody, err := os.ReadFile(filepath.Join(dir, "items.json"))
	if err != nil {
		t.Fatalf("read items json: %v", err)
	}
	if !strings.Contains(string(jsonBody), "\n  {") || !strings.Contains(string(jsonBody), "https://example.com/detail") {
		t.Fatalf("json export body = %s", jsonBody)
	}
	jsonlBody, err := os.ReadFile(filepath.Join(dir, "items.jsonl"))
	if err != nil {
		t.Fatalf("read items jsonl: %v", err)
	}
	if got := strings.Count(string(jsonlBody), "\n"); got != 2 {
		t.Fatalf("jsonl line count = %d, body = %s", got, jsonlBody)
	}
}

type lifecycleResponse struct {
	status int
	body   string
}

type lifecycleSession struct {
	mu        sync.Mutex
	responses map[string]lifecycleResponse
}

func (s *lifecycleSession) Fetch(_ context.Context, request Request) (*goscrapling.Response, error) {
	if strings.HasSuffix(request.URL, "/fail") {
		return nil, errors.New("boom")
	}
	s.mu.Lock()
	response := s.responses[request.URL]
	s.mu.Unlock()
	return goscrapling.NewResponse(strings.NewReader(response.body), goscrapling.ResponseOptions{
		URL:        request.URL,
		StatusCode: response.status,
		Request: goscrapling.RequestMetadata{
			Method:  request.MethodOrDefault(),
			URL:     request.URL,
			Headers: request.Headers,
		},
	})
}

func boolEvent(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
