package goscrapling

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBrowserFetcherContract(t *testing.T) {
	ctx := context.Background()
	engine := &recordingBrowserEngine{
		result: BrowserResult{
			URL:        "https://example.com/rendered",
			StatusCode: http.StatusOK,
			Headers: http.Header{
				"Content-Type": []string{"text/html; charset=utf-8"},
				"X-Engine":     []string{"fake"},
			},
			Body: []byte(`<html><body><article class="quote" data-rendered="true">Loaded by JavaScript</article></body></html>`),
		},
	}
	fetcher := BrowserFetcher{Engine: engine}

	options := BrowserOptions{
		Headers: http.Header{
			"X-Test": []string{"browser"},
		},
		Headless:         false,
		DisableResources: true,
		BlockedDomains:   []string{"ads.example.com", "tracker.example"},
		NetworkIdle:      true,
		LoadDOM:          true,
		Timeout:          2 * time.Second,
		Wait:             25 * time.Millisecond,
		WaitSelector: BrowserWaitSelector{
			Selector: ".quote",
			State:    BrowserWaitVisible,
		},
		Actions: []BrowserAction{
			{Kind: BrowserActionClick, Selector: "#load"},
			{Kind: BrowserActionFill, Selector: "#search", Value: "scrapling"},
		},
	}

	response, err := fetcher.Fetch(ctx, "https://example.com/app", options)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	request := engine.request
	if request.URL != "https://example.com/app" {
		t.Fatalf("expected engine URL, got %q", request.URL)
	}
	if got := request.Headers.Get("X-Test"); got != "browser" {
		t.Fatalf("expected engine header, got %q", got)
	}
	if request.Headless {
		t.Fatal("expected explicit headful option")
	}
	if !request.DisableResources || !request.NetworkIdle || !request.LoadDOM {
		t.Fatalf("expected resource/network/load flags, got %#v", request)
	}
	if !reflect.DeepEqual(request.BlockedDomains, []string{"ads.example.com", "tracker.example"}) {
		t.Fatalf("blocked domains mismatch: %#v", request.BlockedDomains)
	}
	if request.Timeout != 2*time.Second || request.Wait != 25*time.Millisecond {
		t.Fatalf("expected timeout/wait to be forwarded, got timeout=%s wait=%s", request.Timeout, request.Wait)
	}
	if request.WaitSelector.Selector != ".quote" || request.WaitSelector.State != BrowserWaitVisible {
		t.Fatalf("wait selector mismatch: %#v", request.WaitSelector)
	}
	if !reflect.DeepEqual(request.Actions, options.Actions) {
		t.Fatalf("actions mismatch:\nwant: %#v\n got: %#v", options.Actions, request.Actions)
	}

	options.Headers.Set("X-Test", "changed")
	options.BlockedDomains[0] = "changed.example"
	options.Actions[0].Selector = "#changed"
	if got := engine.request.Headers.Get("X-Test"); got != "browser" {
		t.Fatalf("expected request headers to be copied, got %q", got)
	}
	if engine.request.BlockedDomains[0] != "ads.example.com" {
		t.Fatalf("expected blocked domains to be copied, got %#v", engine.request.BlockedDomains)
	}
	if engine.request.Actions[0].Selector != "#load" {
		t.Fatalf("expected actions to be copied, got %#v", engine.request.Actions)
	}

	if response.URL() != "https://example.com/rendered" {
		t.Fatalf("expected final response URL, got %q", response.URL())
	}
	if response.StatusCode() != http.StatusOK {
		t.Fatalf("expected status OK, got %d", response.StatusCode())
	}
	if got := response.Headers().Get("X-Engine"); got != "fake" {
		t.Fatalf("expected response header from engine, got %q", got)
	}
	if got := response.Request().Headers.Get("X-Test"); got != "browser" {
		t.Fatalf("expected response request metadata header, got %q", got)
	}
	first, ok := response.CSS(".quote").First()
	if !ok {
		t.Fatal("expected parsed rendered quote")
	}
	if got := first.Text(); got != "Loaded by JavaScript" {
		t.Fatalf("expected rendered text, got %q", got)
	}
	if got, ok := first.Attr("data-rendered"); !ok || got != "true" {
		t.Fatalf("expected rendered attr, got %q ok=%v", got, ok)
	}
}

type recordingBrowserEngine struct {
	request BrowserRequest
	result  BrowserResult
}

func (e *recordingBrowserEngine) Fetch(_ context.Context, request BrowserRequest) (BrowserResult, error) {
	e.request = request
	return e.result, nil
}

func TestBrowserFetcherRequiresEngine(t *testing.T) {
	_, err := (BrowserFetcher{}).Fetch(context.Background(), "https://example.com", BrowserOptions{})
	if err == nil || !strings.Contains(err.Error(), "browser engine") {
		t.Fatalf("expected missing engine error, got %v", err)
	}
}
