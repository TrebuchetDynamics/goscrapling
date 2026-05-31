package spiders_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestSpiderSessionAdapters(t *testing.T) {
	t.Run("routes static fetcher sessions by sid and preserves request options", func(t *testing.T) {
		var gotMethod, gotBody, gotHeader string
		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotHeader = r.Header.Get("X-Request")
			body, _ := io.ReadAll(r.Body)
			gotBody = string(body)
			_, _ = fmt.Fprint(w, "static:"+gotBody)
		}))
		defer origin.Close()

		staticSession, err := fetchers.NewFetcherSession(fetchers.FetcherSessionOptions{Headers: http.Header{"X-Session": []string{"static"}}})
		if err != nil {
			t.Fatalf("NewFetcherSession returned error: %v", err)
		}
		sessions := spiders.NewSessionManager()
		if err := sessions.Add("static", spiders.NewStaticSessionAdapter(staticSession, spiders.StaticSessionAdapterOptions{}), spiders.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add static session: %v", err)
		}

		crawler := spiders.Crawler{
			Sessions:           sessions,
			ConcurrentRequests: 1,
			DefaultCallback: func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
				return []spiders.Output{spiders.Item(map[string]any{
					"body":   response.Text(),
					"method": response.Response.Request().Method,
				})}, nil
			},
		}
		result, err := crawler.Run(context.Background(), []spiders.Request{{
			URL:     origin.URL,
			SID:     "static",
			Method:  http.MethodPost,
			Body:    []byte("payload"),
			Headers: http.Header{"X-Request": []string{"per-request"}},
		}})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		if gotMethod != http.MethodPost || gotBody != "payload" || gotHeader != "per-request" {
			t.Fatalf("origin saw method/body/header = %q/%q/%q", gotMethod, gotBody, gotHeader)
		}
		if !reflect.DeepEqual(result.Items, []map[string]any{{"body": "static:payload", "method": http.MethodPost}}) {
			t.Fatalf("items = %#v", result.Items)
		}
	})

	t.Run("static proxy rotator is explicit and exposes selected proxy metadata", func(t *testing.T) {
		proxyOne := newAdapterProxy(t, "one")
		defer proxyOne.Close()
		proxyTwo := newAdapterProxy(t, "two")
		defer proxyTwo.Close()
		rotator, err := fetchers.NewProxyRotator([]any{proxyOne.URL, proxyTwo.URL})
		if err != nil {
			t.Fatalf("NewProxyRotator returned error: %v", err)
		}
		staticSession, err := fetchers.NewFetcherSession(fetchers.FetcherSessionOptions{ProxyRotator: rotator})
		if err != nil {
			t.Fatalf("NewFetcherSession returned error: %v", err)
		}
		sessions := spiders.NewSessionManager()
		if err := sessions.Add("rotating", spiders.NewStaticSessionAdapter(staticSession, spiders.StaticSessionAdapterOptions{}), spiders.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add rotating session: %v", err)
		}
		crawler := spiders.Crawler{
			Sessions:           sessions,
			ConcurrentRequests: 1,
			DefaultCallback: func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
				return []spiders.Output{spiders.Item(map[string]any{"body": response.Text(), "proxy": response.Meta["proxy"]})}, nil
			},
		}
		result, err := crawler.Run(context.Background(), []spiders.Request{
			{URL: "http://example.test/first", SID: "rotating"},
			{URL: "http://example.test/second", SID: "rotating"},
		})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		want := []map[string]any{{"body": "one", "proxy": proxyOne.URL}, {"body": "two", "proxy": proxyTwo.URL}}
		if !reflect.DeepEqual(result.Items, want) {
			t.Fatalf("items = %#v, want %#v", result.Items, want)
		}
	})

	t.Run("browser and stealth sessions route by sid with typed options and close through manager", func(t *testing.T) {
		browserEngine := &adapterBrowserEngine{}
		browserSession, err := browser.NewBrowserSession(browser.BrowserSessionOptions{Engine: browserEngine, MaxPages: 1})
		if err != nil {
			t.Fatalf("NewBrowserSession browser returned error: %v", err)
		}
		stealthEngine := &adapterBrowserEngine{}
		stealthSession, err := browser.NewBrowserSession(browser.BrowserSessionOptions{Engine: stealthEngine, MaxPages: 1})
		if err != nil {
			t.Fatalf("NewBrowserSession stealth returned error: %v", err)
		}
		rotator, err := fetchers.NewProxyRotator([]any{"http://proxy-one.example:8080", "http://proxy-two.example:8080"})
		if err != nil {
			t.Fatalf("NewProxyRotator returned error: %v", err)
		}

		sessions := spiders.NewSessionManager()
		if err := sessions.Add("browser", spiders.NewBrowserSessionAdapter(browserSession, spiders.BrowserSessionAdapterOptions{Options: browser.BrowserOptions{NetworkIdle: true}}), spiders.SessionOptions{Default: true}); err != nil {
			t.Fatalf("add browser session: %v", err)
		}
		if err := sessions.Add("stealth", spiders.NewStealthBrowserSessionAdapter(stealthSession, spiders.BrowserSessionAdapterOptions{
			Options:      browser.BrowserOptions{Stealth: browser.BrowserStealthOptions{Enabled: true, GenerateHeaders: true, HideCanvas: true}},
			ProxyRotator: rotator,
		}), spiders.SessionOptions{Lazy: true}); err != nil {
			t.Fatalf("add stealth session: %v", err)
		}

		crawler := spiders.Crawler{
			Sessions:           sessions,
			ConcurrentRequests: 1,
			DefaultCallback: func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
				return []spiders.Output{spiders.Item(map[string]any{"url": response.URL(), "proxy": response.Meta["proxy"]})}, nil
			},
		}
		browserReq := spiders.WithBrowserRequestOptions(spiders.Request{URL: "https://example.com/app", SID: "browser"}, browser.BrowserOptions{
			Headers:      http.Header{"X-Browser": []string{"request"}},
			Wait:         25 * time.Millisecond,
			WaitSelector: browser.BrowserWaitSelector{Selector: "#ready", State: browser.BrowserWaitVisible},
		})
		result, err := crawler.Run(context.Background(), []spiders.Request{
			browserReq,
			{URL: "https://example.com/protected", SID: "stealth"},
			{URL: "https://example.com/protected-two", SID: "stealth"},
		})
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
		if len(result.Items) != 3 {
			t.Fatalf("items = %#v", result.Items)
		}
		browserRequests := browserEngine.Requests()
		if len(browserRequests) != 1 {
			t.Fatalf("browser engine requests = %d", len(browserRequests))
		}
		if !browserRequests[0].NetworkIdle || browserRequests[0].Wait != 25*time.Millisecond || browserRequests[0].WaitSelector.Selector != "#ready" || browserRequests[0].Headers.Get("X-Browser") != "request" {
			t.Fatalf("browser request options = %#v", browserRequests[0])
		}
		stealthRequests := stealthEngine.Requests()
		if len(stealthRequests) != 2 {
			t.Fatalf("stealth engine requests = %d", len(stealthRequests))
		}
		if !stealthRequests[0].Stealth.Enabled || !stealthRequests[0].Stealth.GenerateHeaders || !stealthRequests[0].Stealth.HideCanvas || stealthRequests[0].Headers.Get("User-Agent") == "" {
			t.Fatalf("stealth request did not preserve explicit stealth controls: %#v", stealthRequests[0])
		}
		if stealthRequests[0].Proxy.Server != "http://proxy-one.example:8080" || stealthRequests[1].Proxy.Server != "http://proxy-two.example:8080" {
			t.Fatalf("stealth proxy rotation = %#v / %#v", stealthRequests[0].Proxy, stealthRequests[1].Proxy)
		}
		if result.Items[1]["proxy"] != "http://proxy-one.example:8080" || result.Items[2]["proxy"] != "http://proxy-two.example:8080" {
			t.Fatalf("browser proxy meta items = %#v", result.Items)
		}
		if !browserSession.Stats().Closed || !stealthSession.Stats().Closed {
			t.Fatalf("sessions not closed: browser=%#v stealth=%#v", browserSession.Stats(), stealthSession.Stats())
		}
	})
}

func newAdapterProxy(t *testing.T, name string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, name)
	}))
}

type adapterBrowserEngine struct {
	mu       sync.Mutex
	requests []browser.BrowserRequest
}

func (e *adapterBrowserEngine) Fetch(_ context.Context, request browser.BrowserRequest) (browser.BrowserResult, error) {
	e.mu.Lock()
	e.requests = append(e.requests, request)
	e.mu.Unlock()
	return browser.BrowserResult{
		URL:        request.URL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Body:       []byte("<html><body>" + request.URL + "</body></html>"),
	}, nil
}

func (e *adapterBrowserEngine) Requests() []browser.BrowserRequest {
	e.mu.Lock()
	defer e.mu.Unlock()
	requests := make([]browser.BrowserRequest, len(e.requests))
	copy(requests, e.requests)
	return requests
}
