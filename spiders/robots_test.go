package spiders

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/goscrapling"
)

func TestRobotsTxtManager(t *testing.T) {
	t.Run("merges duplicate matching user-agent groups", func(t *testing.T) {
		parser := parseRobotsTxt(strings.Join([]string{
			"User-agent: GoodBot",
			"Disallow: /private",
			"",
			"User-agent: OtherBot",
			"Disallow: /other",
			"",
			"User-agent: GoodBot",
			"Allow: /private/public",
			"Crawl-delay: 2",
		}, "\n"))

		if parser.canFetch("https://example.com/private/report", "GoodBot/1.0") {
			t.Fatal("merged GoodBot groups should deny /private")
		}
		if !parser.canFetch("https://example.com/private/public/index", "GoodBot/1.0") {
			t.Fatal("merged GoodBot groups should allow longer /private/public rule")
		}
		if got := parser.delayDirectives("GoodBot/1.0").CrawlDelay; got != 2*time.Second {
			t.Fatalf("merged GoodBot crawl delay = %s, want 2s", got)
		}
	})

	t.Run("parses allow deny and delay directives from local fixtures", func(t *testing.T) {
		fetcher := newRobotsFixtureFetcher(t)
		manager := NewRobotsTxtManager(fetcher.Fetch)
		ctx := context.Background()

		if err := manager.Prefetch(ctx, []string{
			"https://example.com/start",
			"https://example.com/other",
			"https://other.example/start",
		}, "robot-sid"); err != nil {
			t.Fatalf("Prefetch returned error: %v", err)
		}
		if got := fetcher.URLs(); !reflect.DeepEqual(got, []string{"https://example.com/robots.txt", "https://other.example/robots.txt"}) {
			t.Fatalf("prefetched robots URLs = %#v", got)
		}
		if got := fetcher.SIDs(); !reflect.DeepEqual(got, []string{"robot-sid", "robot-sid"}) {
			t.Fatalf("prefetch SIDs = %#v", got)
		}

		allowed, err := manager.CanFetch(ctx, "https://example.com/private/report", "robot-sid", "GoScraplingBot")
		if err != nil {
			t.Fatalf("CanFetch private returned error: %v", err)
		}
		if allowed {
			t.Fatal("GoScraplingBot should be denied by /private")
		}
		allowed, err = manager.CanFetch(ctx, "https://example.com/private/public/index", "robot-sid", "GoScraplingBot")
		if err != nil {
			t.Fatalf("CanFetch public returned error: %v", err)
		}
		if !allowed {
			t.Fatal("longer Allow rule should permit /private/public")
		}
		allowed, err = manager.CanFetch(ctx, "https://example.com/tmp/cache", "robot-sid", "OtherBot")
		if err != nil {
			t.Fatalf("CanFetch wildcard returned error: %v", err)
		}
		if allowed {
			t.Fatal("wildcard user-agent should be denied by /tmp")
		}

		directives, err := manager.DelayDirectives(ctx, "https://example.com/private/public/index", "robot-sid", "GoScraplingBot")
		if err != nil {
			t.Fatalf("DelayDirectives returned error: %v", err)
		}
		if directives.CrawlDelay != 1500*time.Millisecond {
			t.Fatalf("crawl delay = %s, want 1.5s", directives.CrawlDelay)
		}
		if directives.RequestRate == nil || directives.RequestRate.Requests != 4 || directives.RequestRate.Period != 8*time.Second {
			t.Fatalf("request rate = %#v, want 4 per 8s", directives.RequestRate)
		}
		if got := directives.EffectiveDelay(500 * time.Millisecond); got != 2*time.Second {
			t.Fatalf("effective delay = %s, want max of configured/crawl-delay/request-rate = 2s", got)
		}

		_ = fetcher.URLs()
		if got := fetcher.Count(); got != 2 {
			t.Fatalf("robots cache miss count = %d, want prefetch-only two fetches", got)
		}
	})

	t.Run("crawler filters disallowed requests, applies robots delay, and records stats", func(t *testing.T) {
		session := &robotsCrawlerSession{fixtures: map[string]string{
			"https://crawl.example/robots.txt": fixtureRobots(t, "crawl.example"),
			"https://crawl.example/allowed":    "<html>allowed</html>",
		}}
		sessions := NewSessionManager()
		if err := sessions.Add("default", session, SessionOptions{Default: true}); err != nil {
			t.Fatalf("add session: %v", err)
		}

		var delays []time.Duration
		crawler := Crawler{
			Sessions:           sessions,
			RobotsTxtObey:      true,
			RobotsUserAgent:    "GoodBot",
			DownloadDelay:      100 * time.Millisecond,
			ConcurrentRequests: 1,
			sleep: func(_ context.Context, delay time.Duration) error {
				delays = append(delays, delay)
				return nil
			},
			DefaultCallback: func(_ context.Context, response Response) ([]Output, error) {
				return []Output{Item(map[string]any{"url": response.URL()})}, nil
			},
		}
		result, err := crawler.Run(context.Background(), []Request{
			{URL: "https://crawl.example/allowed", Headers: http.Header{"User-Agent": []string{"GoodBot"}}},
			{URL: "https://crawl.example/blocked", Headers: http.Header{"User-Agent": []string{"GoodBot"}}},
		})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		if result.Stats.Requests != 1 || result.Stats.RobotsDisallowed != 1 || result.Stats.Items != 1 {
			t.Fatalf("stats = %#v, want one fetched, one robots-disallowed, one item", result.Stats)
		}
		if !reflect.DeepEqual(delays, []time.Duration{500 * time.Millisecond}) {
			t.Fatalf("delays = %#v, want robots request-rate delay only for allowed fetch", delays)
		}
		if got := session.URLs(); !reflect.DeepEqual(got, []string{"https://crawl.example/allowed", "https://crawl.example/robots.txt"}) {
			t.Fatalf("fetched URLs = %#v", got)
		}
		if !reflect.DeepEqual(result.Items, []map[string]any{{"url": "https://crawl.example/allowed"}}) {
			t.Fatalf("items = %#v", result.Items)
		}
	})
}

type robotsFixtureFetcher struct {
	t        *testing.T
	mu       sync.Mutex
	urls     []string
	sids     []string
	fixtures map[string]string
}

func newRobotsFixtureFetcher(t *testing.T) *robotsFixtureFetcher {
	t.Helper()
	return &robotsFixtureFetcher{
		t: t,
		fixtures: map[string]string{
			"example.com":   fixtureRobots(t, "example.com"),
			"other.example": fixtureRobots(t, "other.example"),
		},
	}
}

func (f *robotsFixtureFetcher) Fetch(_ context.Context, robotsURL, sid string) (Response, error) {
	f.mu.Lock()
	f.urls = append(f.urls, robotsURL)
	f.sids = append(f.sids, sid)
	f.mu.Unlock()

	parsed, err := url.Parse(robotsURL)
	if err != nil {
		return Response{}, err
	}
	body := f.fixtures[parsed.Host]
	response, err := goscrapling.NewResponse(strings.NewReader(body), goscrapling.ResponseOptions{
		URL:        robotsURL,
		StatusCode: http.StatusOK,
		Request: goscrapling.RequestMetadata{
			Method: http.MethodGet,
			URL:    robotsURL,
		},
	})
	if err != nil {
		return Response{}, err
	}
	return Response{Response: response, Request: Request{URL: robotsURL, SID: sid}}, nil
}

func (f *robotsFixtureFetcher) URLs() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	urls := append([]string(nil), f.urls...)
	sort.Strings(urls)
	return urls
}

func (f *robotsFixtureFetcher) SIDs() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	sids := append([]string(nil), f.sids...)
	sort.Strings(sids)
	return sids
}

func (f *robotsFixtureFetcher) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.urls)
}

type robotsCrawlerSession struct {
	mu       sync.Mutex
	fixtures map[string]string
	requests []Request
}

func (s *robotsCrawlerSession) Fetch(_ context.Context, request Request) (*goscrapling.Response, error) {
	s.mu.Lock()
	s.requests = append(s.requests, request.clone())
	s.mu.Unlock()

	body := s.fixtures[request.URL]
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

func (s *robotsCrawlerSession) URLs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	urls := make([]string, 0, len(s.requests))
	for _, request := range s.requests {
		urls = append(urls, request.URL)
	}
	sort.Strings(urls)
	return urls
}

func fixtureRobots(t *testing.T, host string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "testdata", "spiders", "robots", host+".txt"))
	if err != nil {
		t.Fatalf("read robots fixture for %s: %v", host, err)
	}
	return string(body)
}
