package spider_test

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/spider"
)

func TestSpider(t *testing.T) {
	t.Run("scheduler prioritizes requests and deduplicates fingerprints", func(t *testing.T) {
		scheduler := spider.NewScheduler(spider.SchedulerOptions{})

		low := spider.Request{URL: "https://example.com/list#ignored", Priority: 1}
		high := spider.Request{URL: "https://example.com/detail", Priority: 10}
		duplicateLow := spider.Request{URL: "https://example.com/list", Priority: 99}
		query := spider.Request{URL: "https://example.com/search?b=2&a=1", Priority: 5}
		duplicateQuery := spider.Request{URL: "https://example.com/search?a=1&b=2", Priority: 6}

		if ok, err := scheduler.Enqueue(low); err != nil || !ok {
			t.Fatalf("enqueue low ok=%v err=%v", ok, err)
		}
		if ok, err := scheduler.Enqueue(high); err != nil || !ok {
			t.Fatalf("enqueue high ok=%v err=%v", ok, err)
		}
		if ok, err := scheduler.Enqueue(duplicateLow); err != nil || ok {
			t.Fatalf("duplicate enqueue ok=%v err=%v, want dropped without error", ok, err)
		}
		if ok, err := scheduler.Enqueue(query); err != nil || !ok {
			t.Fatalf("enqueue query ok=%v err=%v", ok, err)
		}
		if ok, err := scheduler.Enqueue(duplicateQuery); err != nil || ok {
			t.Fatalf("duplicate query enqueue ok=%v err=%v, want dropped without error", ok, err)
		}

		first, ok := scheduler.Dequeue()
		if !ok {
			t.Fatal("expected first request")
		}
		if first.URL != high.URL {
			t.Fatalf("expected high priority first, got %q", first.URL)
		}
		second, ok := scheduler.Dequeue()
		if !ok {
			t.Fatal("expected second request")
		}
		if second.URL != query.URL {
			t.Fatalf("expected query priority second, got %q", second.URL)
		}
		third, ok := scheduler.Dequeue()
		if !ok {
			t.Fatal("expected third request")
		}
		if third.URL != low.URL {
			t.Fatalf("expected low priority third, got %q", third.URL)
		}
		if _, ok := scheduler.Dequeue(); ok {
			t.Fatal("expected empty scheduler")
		}
	})

	t.Run("crawler routes sessions, follows requests, and collects results", func(t *testing.T) {
		ctx := context.Background()
		listSession := &fakeSession{
			responses: map[string]string{
				"https://example.com/list": `<html><body><a class="detail" href="/detail">Detail</a></body></html>`,
			},
		}
		detailSession := &fakeSession{
			responses: map[string]string{
				"https://example.com/detail": `<html><body><h1>Detail title</h1></body></html>`,
			},
		}

		sessions := spider.NewSessionManager()
		if err := sessions.Add("list", listSession, spider.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add list session: %v", err)
		}
		if err := sessions.Add("detail", detailSession, spider.SessionOptions{Lazy: true}); err != nil {
			t.Fatalf("add detail session: %v", err)
		}

		var parseDetail spider.Callback
		parseDetail = func(_ context.Context, response spider.Response) ([]spider.Output, error) {
			title, ok := response.CSS("h1").First()
			if !ok {
				t.Fatal("expected detail title")
			}
			return []spider.Output{
				spider.Item(map[string]any{
					"title":    title.Text(),
					"category": response.Meta["category"],
					"sid":      response.Request.SID,
				}),
			}, nil
		}

		parseList := func(_ context.Context, response spider.Response) ([]spider.Output, error) {
			follow, err := response.Follow("/detail", spider.FollowOptions{
				SID:      "detail",
				Callback: parseDetail,
				Priority: spider.Priority(5),
				Meta: map[string]any{
					"category": "docs",
				},
			})
			if err != nil {
				return nil, err
			}
			return []spider.Output{spider.Next(follow)}, nil
		}

		crawler := spider.Crawler{
			Sessions:        sessions,
			DefaultCallback: parseList,
		}
		result, err := crawler.Run(ctx, []spider.Request{
			{URL: "https://example.com/list", SID: "list"},
			{URL: "https://example.com/list", SID: "list"},
		})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}

		if !reflect.DeepEqual(result.Items, []map[string]any{{
			"title":    "Detail title",
			"category": "docs",
			"sid":      "detail",
		}}) {
			t.Fatalf("items mismatch: %#v", result.Items)
		}
		if result.Stats.Requests != 2 {
			t.Fatalf("expected 2 fetched requests after dedupe, got %d", result.Stats.Requests)
		}
		if result.Stats.Items != 1 {
			t.Fatalf("expected 1 item, got %d", result.Stats.Items)
		}
		if result.Stats.Skipped != 1 {
			t.Fatalf("expected 1 skipped duplicate request, got %d", result.Stats.Skipped)
		}
		if got := result.Stats.Sessions["list"]; got != 1 {
			t.Fatalf("expected list session count 1, got %d", got)
		}
		if got := result.Stats.Sessions["detail"]; got != 1 {
			t.Fatalf("expected detail session count 1, got %d", got)
		}
		if listSession.starts != 1 || listSession.closes != 1 {
			t.Fatalf("expected eager list session start/close once, got starts=%d closes=%d", listSession.starts, listSession.closes)
		}
		if detailSession.starts != 1 || detailSession.closes != 1 {
			t.Fatalf("expected lazy detail session start/close once, got starts=%d closes=%d", detailSession.starts, detailSession.closes)
		}
		if len(listSession.requests) != 1 || len(detailSession.requests) != 1 {
			t.Fatalf("expected one request per session, got list=%d detail=%d", len(listSession.requests), len(detailSession.requests))
		}
		if got := detailSession.requests[0].Headers.Get("Referer"); got != "https://example.com/list" {
			t.Fatalf("expected referer flow header, got %q", got)
		}
	})
}

type fakeSession struct {
	responses map[string]string
	requests  []spider.Request
	starts    int
	closes    int
}

func (s *fakeSession) Start(context.Context) error {
	s.starts++
	return nil
}

func (s *fakeSession) Close(context.Context) error {
	s.closes++
	return nil
}

func (s *fakeSession) Fetch(_ context.Context, request spider.Request) (*goscrapling.Response, error) {
	s.requests = append(s.requests, request)
	body := s.responses[request.URL]
	return goscrapling.NewResponse(strings.NewReader(body), goscrapling.ResponseOptions{
		URL:        request.URL,
		StatusCode: http.StatusOK,
		Request: goscrapling.RequestMetadata{
			Method:  request.MethodOrDefault(),
			URL:     request.URL,
			Headers: request.Headers,
		},
	})
}
