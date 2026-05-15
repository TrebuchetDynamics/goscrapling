package spider_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/spider"
)

func TestSpiderAllowedDomains(t *testing.T) {
	ctx := context.Background()
	session := &fakeSession{
		responses: map[string]string{
			"https://example.com/list":        `<html><body>list</body></html>`,
			"https://example.com/detail":      `<html><body>detail</body></html>`,
			"https://blog.example.com/detail": `<html><body>blog detail</body></html>`,
		},
	}

	sessions := spider.NewSessionManager()
	if err := sessions.Add("default", session, spider.SessionOptions{Default: true}); err != nil {
		t.Fatalf("add session: %v", err)
	}

	parseDetail := func(_ context.Context, response spider.Response) ([]spider.Output, error) {
		return []spider.Output{
			spider.Item(map[string]any{"url": response.URL()}),
		}, nil
	}
	parseList := func(_ context.Context, response spider.Response) ([]spider.Output, error) {
		sameDomain, err := response.Follow("/detail", spider.FollowOptions{Callback: parseDetail})
		if err != nil {
			return nil, err
		}
		subdomain, err := response.Follow("https://blog.example.com/detail", spider.FollowOptions{Callback: parseDetail})
		if err != nil {
			return nil, err
		}
		offsite, err := response.Follow("https://outside.example.net/detail", spider.FollowOptions{Callback: parseDetail})
		if err != nil {
			return nil, err
		}
		suffixTrap, err := response.Follow("https://notexample.com/detail", spider.FollowOptions{Callback: parseDetail})
		if err != nil {
			return nil, err
		}
		nestedSuffixTrap, err := response.Follow("https://example.com.evil.test/detail", spider.FollowOptions{Callback: parseDetail})
		if err != nil {
			return nil, err
		}

		return []spider.Output{
			spider.Next(sameDomain),
			spider.Next(subdomain),
			spider.Next(offsite),
			spider.Next(suffixTrap),
			spider.Next(nestedSuffixTrap),
		}, nil
	}

	crawler := spider.Crawler{
		Sessions:        sessions,
		DefaultCallback: parseList,
		AllowedDomains:  []string{"example.com"},
	}
	result, err := crawler.Run(ctx, []spider.Request{
		{URL: "https://example.com/list"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	gotFetchedURLs := make([]string, 0, len(session.requests))
	for _, request := range session.requests {
		gotFetchedURLs = append(gotFetchedURLs, request.URL)
	}
	wantFetchedURLs := []string{
		"https://example.com/list",
		"https://example.com/detail",
		"https://blog.example.com/detail",
	}
	sort.Strings(gotFetchedURLs)
	sort.Strings(wantFetchedURLs)
	if !reflect.DeepEqual(gotFetchedURLs, wantFetchedURLs) {
		t.Fatalf("fetched URLs mismatch:\ngot  %#v\nwant %#v", gotFetchedURLs, wantFetchedURLs)
	}
	if result.Stats.Requests != 3 {
		t.Fatalf("requests = %d, want 3 fetched requests", result.Stats.Requests)
	}
	if result.Stats.OffsiteRequests != 3 {
		t.Fatalf("offsite requests = %d, want 3 filtered requests", result.Stats.OffsiteRequests)
	}
	if result.Stats.Skipped != 0 {
		t.Fatalf("skipped = %d, want 0 duplicate skips", result.Stats.Skipped)
	}

	gotItemURLs := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		gotItemURLs = append(gotItemURLs, item["url"].(string))
	}
	wantItemURLs := []string{
		"https://example.com/detail",
		"https://blog.example.com/detail",
	}
	sort.Strings(gotItemURLs)
	sort.Strings(wantItemURLs)
	if !reflect.DeepEqual(gotItemURLs, wantItemURLs) {
		t.Fatalf("item URLs mismatch:\ngot  %#v\nwant %#v", gotItemURLs, wantItemURLs)
	}
}
