package spiders

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
)

func TestSpiderCheckpoint(t *testing.T) {
	t.Run("manager atomically saves, loads, restores, and cleans up scheduler state", func(t *testing.T) {
		dir := t.TempDir()
		manager := NewCheckpointManager(dir)
		scheduler := NewScheduler(SchedulerOptions{})
		if ok, err := scheduler.Enqueue(Request{URL: "https://example.com/low", Priority: 1}); err != nil || !ok {
			t.Fatalf("enqueue low ok=%v err=%v", ok, err)
		}
		if ok, err := scheduler.Enqueue(Request{URL: "https://example.com/high", Priority: 5}); err != nil || !ok {
			t.Fatalf("enqueue high ok=%v err=%v", ok, err)
		}

		snapshot := scheduler.Snapshot()
		if err := manager.Save(snapshot); err != nil {
			t.Fatalf("Save returned error: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "checkpoint.json")); err != nil {
			t.Fatalf("checkpoint file missing: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "checkpoint.json.tmp")); !os.IsNotExist(err) {
			t.Fatalf("temp checkpoint file was not cleaned up, stat err=%v", err)
		}

		loaded, ok, err := manager.Load()
		if err != nil || !ok {
			t.Fatalf("Load ok=%v err=%v", ok, err)
		}
		if got := requestURLs(loaded.Requests); !reflect.DeepEqual(got, []string{"https://example.com/high", "https://example.com/low"}) {
			t.Fatalf("loaded request order = %#v", got)
		}

		restored := NewScheduler(SchedulerOptions{})
		restored.Restore(loaded)
		first, ok := restored.Dequeue()
		if !ok || first.URL != "https://example.com/high" {
			t.Fatalf("first restored request = %#v ok=%v", first, ok)
		}
		if queued, err := restored.Enqueue(Request{URL: "https://example.com/low"}); err != nil || queued {
			t.Fatalf("seen restoration failed queued=%v err=%v", queued, err)
		}
		if err := manager.Cleanup(); err != nil {
			t.Fatalf("Cleanup returned error: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "checkpoint.json")); !os.IsNotExist(err) {
			t.Fatalf("checkpoint file still exists, stat err=%v", err)
		}
	})

	t.Run("crawler pause saves pending work and later resumes without replaying completed starts", func(t *testing.T) {
		dir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		firstSession := &checkpointSession{}
		firstSessions := NewSessionManager()
		if err := firstSessions.Add("default", firstSession, SessionOptions{Default: true}); err != nil {
			t.Fatalf("add first session: %v", err)
		}
		firstCrawler := Crawler{
			Sessions:           firstSessions,
			CheckpointDir:      dir,
			ConcurrentRequests: 1,
			DefaultCallback: func(_ context.Context, response Response) ([]Output, error) {
				if strings.HasSuffix(response.Request.URL, "/first") {
					cancel()
				}
				return []Output{Item(map[string]any{"url": response.Request.URL})}, nil
			},
		}
		paused, err := firstCrawler.Run(ctx, []Request{
			{URL: "https://example.com/first", Priority: 10},
			{URL: "https://example.com/second", Priority: 1},
		})
		if err != nil {
			t.Fatalf("paused Run returned error: %v", err)
		}
		if !paused.Paused || paused.Stats.Requests != 1 {
			t.Fatalf("paused result = %#v", paused)
		}
		if got := requestURLs(firstSession.requests); !reflect.DeepEqual(got, []string{"https://example.com/first"}) {
			t.Fatalf("first run requests = %#v", got)
		}

		secondSession := &checkpointSession{}
		secondSessions := NewSessionManager()
		if err := secondSessions.Add("default", secondSession, SessionOptions{Default: true}); err != nil {
			t.Fatalf("add second session: %v", err)
		}
		secondCrawler := Crawler{
			Sessions:           secondSessions,
			CheckpointDir:      dir,
			ConcurrentRequests: 1,
			DefaultCallback: func(_ context.Context, response Response) ([]Output, error) {
				return []Output{Item(map[string]any{"url": response.Request.URL})}, nil
			},
		}
		resumed, err := secondCrawler.Run(context.Background(), []Request{{URL: "https://example.com/first", Priority: 10}})
		if err != nil {
			t.Fatalf("resumed Run returned error: %v", err)
		}
		if resumed.Paused {
			t.Fatalf("resumed result should not be paused: %#v", resumed)
		}
		if got := requestURLs(secondSession.requests); !reflect.DeepEqual(got, []string{"https://example.com/second"}) {
			t.Fatalf("resumed requests = %#v", got)
		}
		if _, err := os.Stat(filepath.Join(dir, "checkpoint.json")); !os.IsNotExist(err) {
			t.Fatalf("checkpoint file should be cleaned after successful resume, stat err=%v", err)
		}
	})
}

type checkpointSession struct {
	mu       sync.Mutex
	requests []Request
}

func (s *checkpointSession) Fetch(_ context.Context, request Request) (*goscrapling.Response, error) {
	s.mu.Lock()
	s.requests = append(s.requests, request.clone())
	s.mu.Unlock()
	return goscrapling.NewResponse(strings.NewReader(request.URL), goscrapling.ResponseOptions{
		URL:        request.URL,
		StatusCode: http.StatusOK,
		Request: goscrapling.RequestMetadata{
			Method: request.MethodOrDefault(),
			URL:    request.URL,
		},
	})
}

func requestURLs(requests []Request) []string {
	urls := make([]string, 0, len(requests))
	for _, request := range requests {
		urls = append(urls, request.URL)
	}
	return urls
}
