package browser

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrowserAdapter(t *testing.T) {
	chromePath, ok := testChromeExecutable()
	if !ok {
		t.Skip("real browser adapter test requires Chrome or Chromium; set GOSCRAPLING_CHROME or GOSCRAPLING_CHROME_AUTO=1 to run it")
	}

	fixture, err := os.ReadFile(filepath.Join("..", "..", "testdata", "browser", "dynamic.html"))
	if err != nil {
		t.Fatalf("read browser fixture: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dynamic":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			body := strings.ReplaceAll(string(fixture), "{{REQUEST_HEADER}}", r.Header.Get("X-Browser-Test"))
			fmt.Fprint(w, body)
		case "/never":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<!doctype html><html><body><main id="app"></main></body></html>`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	fetcher := BrowserFetcher{
		Engine: NewChromedpBrowserEngine(ChromedpBrowserOptions{
			ExecutablePath: chromePath,
		}),
	}

	response, err := fetcher.Fetch(context.Background(), server.URL+"/dynamic", BrowserOptions{
		Headers: http.Header{
			"X-Browser-Test": []string{"real-browser"},
		},
		Headless: true,
		LoadDOM:  true,
		Timeout:  5 * time.Second,
		WaitSelector: BrowserWaitSelector{
			Selector: "#rendered",
			State:    BrowserWaitVisible,
		},
	})
	if err != nil {
		t.Fatalf("BrowserFetcher.Fetch returned error: %v", err)
	}
	if response.StatusCode() != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode())
	}
	if !strings.HasPrefix(response.URL(), server.URL+"/dynamic") {
		t.Fatalf("response URL = %q, want dynamic fixture URL", response.URL())
	}
	if got := response.Request().Headers.Get("X-Browser-Test"); got != "real-browser" {
		t.Fatalf("request metadata header = %q, want real-browser", got)
	}
	if got := response.Headers().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content type = %q, want text/html", got)
	}

	rendered, ok := response.CSS("#rendered").First()
	if !ok {
		t.Fatal("expected JavaScript-rendered article")
	}
	if got := rendered.Text(); got != "Rendered by browser JavaScript" {
		t.Fatalf("rendered text = %q, want JavaScript output", got)
	}
	if got, ok := rendered.Attr("data-header"); !ok || got != "real-browser" {
		t.Fatalf("rendered data-header = %q ok=%v, want real-browser", got, ok)
	}

	_, err = fetcher.Fetch(context.Background(), server.URL+"/never", BrowserOptions{
		Headless: true,
		Timeout:  2 * time.Second,
		WaitSelector: BrowserWaitSelector{
			Selector: "#missing",
			State:    BrowserWaitVisible,
		},
	})
	if err == nil {
		t.Fatal("expected timeout waiting for missing selector")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(strings.ToLower(err.Error()), "context deadline") {
		t.Fatalf("timeout error = %v, want context deadline", err)
	}
}

func testChromeExecutable() (string, bool) {
	if path := os.Getenv("GOSCRAPLING_CHROME"); path != "" {
		return path, true
	}
	if os.Getenv("GOSCRAPLING_CHROME_AUTO") != "1" {
		return "", false
	}
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, true
		}
	}
	return "", false
}
