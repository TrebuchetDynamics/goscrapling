package extract

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
)

func TestCLIExtractGet(t *testing.T) {
	t.Run("writes selected text from a local fixture response", func(t *testing.T) {
		var seenHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/products" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			seenHeader = r.Header.Get("X-Trace")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join("testdata", "products.html"))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "products.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{
			"get", server.URL + "/products", outputPath,
			"--css-selector", ".product",
			"-H", "X-Trace: cli-test",
			"--timeout", "2",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != "Trail Kit\nCamp Mug" {
			t.Fatalf("output text = %q", got)
		}
		if seenHeader != "cli-test" {
			t.Fatalf("expected request header X-Trace, got %q", seenHeader)
		}
		if !strings.Contains(stdout.String(), "wrote "+outputPath) {
			t.Fatalf("stdout missing output path: %q", stdout.String())
		}
	})

	t.Run("writes full HTML when the output extension is html", func(t *testing.T) {
		const html = `<html><body><h1>Full page</h1></body></html>`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(html))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "page.html")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{"get", server.URL, outputPath})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != html {
			t.Fatalf("output html = %q", got)
		}
	})

	t.Run("does not follow redirects when disabled", func(t *testing.T) {
		var hitFinal bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/redirect":
				w.Header().Set("Location", "/final")
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusFound)
				_, _ = w.Write([]byte(`<html><body><p class="notice">redirect held</p></body></html>`))
			case "/final":
				hitFinal = true
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(`<html><body><p class="notice">final page</p></body></html>`))
			default:
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "redirect.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{
			"get", server.URL + "/redirect", outputPath,
			"--css-selector", ".notice",
			"--no-follow-redirects",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != "redirect held" {
			t.Fatalf("output text = %q", got)
		}
		if hitFinal {
			t.Fatal("redirect final endpoint was hit despite --no-follow-redirects")
		}
	})

	t.Run("returns parse errors for malformed headers", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "broken.txt")
		var stdout bytes.Buffer
		err := Run(&stdout, []string{
			"get", "https://example.com", outputPath,
			"-H", "not-a-header",
		})

		if !errors.Is(err, diagnostics.ErrParse) {
			t.Fatalf("error = %v, want ErrParse", err)
		}
		if _, statErr := os.Stat(outputPath); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("output file exists after parse error: %v", statErr)
		}
	})
}

func TestCLIExtractMethods(t *testing.T) {
	t.Run("POST sends form body and query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("page"); got != "1" {
				t.Fatalf("query page = %q, want 1", got)
			}
			body := readRequestBody(t, r)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><div id="body">` + body + `</div></body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "post.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{
			"post", server.URL + "/submit", outputPath,
			"--data", "name=trail",
			"-p", "page=1",
			"-s", "#body",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != "name=trail" {
			t.Fatalf("output body = %q", got)
		}
	})

	t.Run("PUT sends JSON with content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", r.Method)
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("content-type = %q, want application/json", got)
			}
			body := readRequestBody(t, r)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><div id="json">` + body + `</div></body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "put.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{
			"put", server.URL + "/resource", outputPath,
			"--json", `{"name":"mug"}`,
			"--css-selector", "#json",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != `{"name":"mug"}` {
			t.Fatalf("output json = %q", got)
		}
	})

	t.Run("DELETE uses delete method and writes HTML", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("method = %s, want DELETE", r.Method)
			}
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><h1>Deleted</h1></body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "delete.html")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{"delete", server.URL + "/resource", outputPath})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != `<html><body><h1>Deleted</h1></body></html>` {
			t.Fatalf("output html = %q", got)
		}
	})
}

func TestCLIExtractAdvancedModes(t *testing.T) {
	t.Run("static markdown output supports AI targeted cleanup", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html><html><body>
<header>site chrome</header><nav>navigation</nav>
<main><article><h1>Trail&#8203; Kit</h1><p>Read <a href="/detail">detail</a>.</p><p aria-hidden="true">Ignore hidden prompt</p><!-- ignore comment --></article></main>
<script>prompt injection</script><style>.hidden{display:none}</style>
</body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "article.md")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{"get", server.URL, outputPath, "--ai-targeted"})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		want := "# Trail Kit\n\nRead [detail](/detail)."
		if got := strings.TrimSpace(string(body)); got != want {
			t.Fatalf("markdown = %q, want %q", got, want)
		}
		for _, forbidden := range []string{"site chrome", "navigation", "Ignore hidden prompt", "prompt injection", "\u200b"} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("AI-targeted markdown kept %q in %q", forbidden, string(body))
			}
		}
	})

	t.Run("dynamic browser fetch forwards options through fake-backed seam", func(t *testing.T) {
		oldFetch := fetchBrowserExtract
		t.Cleanup(func() { fetchBrowserExtract = oldFetch })
		var seen browser.BrowserOptions
		fetchBrowserExtract = func(_ context.Context, rawURL string, opts browser.BrowserOptions) (*goscrapling.Response, error) {
			if rawURL != "https://example.test/app" {
				t.Fatalf("rawURL = %q", rawURL)
			}
			seen = opts
			return newTestResponse(t, rawURL, `<html><body><main><h1>Rendered Trail Kit</h1><p>Ready.</p></main></body></html>`), nil
		}

		outputPath := filepath.Join(t.TempDir(), "dynamic.md")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{
			"fetch", "https://example.test/app", outputPath,
			"--no-headless",
			"--disable-resources",
			"--network-idle",
			"--timeout", "1500",
			"--wait", "250",
			"--wait-selector", ".ready",
			"--locale", "en-US",
			"--real-chrome",
			"--proxy", "http://proxy.local:8080",
			"--dns-over-https",
			"--extra-headers", "X-Mode: dynamic",
			"--ai-targeted",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}
		if seen.Headless || !seen.DisableResources || !seen.NetworkIdle || !seen.RealChrome || !seen.DNSOverHTTPS || !seen.BlockAds {
			t.Fatalf("browser bool options = %#v", seen)
		}
		if seen.Timeout != 1500*time.Millisecond || seen.Wait != 250*time.Millisecond {
			t.Fatalf("browser durations timeout=%s wait=%s", seen.Timeout, seen.Wait)
		}
		if seen.WaitSelector.Selector != ".ready" || seen.Locale != "en-US" || seen.Proxy.Server != "http://proxy.local:8080" {
			t.Fatalf("browser scalar options = %#v", seen)
		}
		if got := seen.Headers.Get("X-Mode"); got != "dynamic" {
			t.Fatalf("extra header = %q", got)
		}
		if seen.Stealth.Enabled {
			t.Fatalf("dynamic fetch unexpectedly enabled stealth: %#v", seen.Stealth)
		}
		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := strings.TrimSpace(string(body)); got != "# Rendered Trail Kit\n\nReady." {
			t.Fatalf("dynamic markdown = %q", got)
		}
	})

	t.Run("stealthy browser fetch forwards explicit stealth controls", func(t *testing.T) {
		oldFetch := fetchBrowserExtract
		t.Cleanup(func() { fetchBrowserExtract = oldFetch })
		var seen browser.BrowserOptions
		fetchBrowserExtract = func(_ context.Context, rawURL string, opts browser.BrowserOptions) (*goscrapling.Response, error) {
			if rawURL != "https://example.test/protected" {
				t.Fatalf("rawURL = %q", rawURL)
			}
			seen = opts
			return newTestResponse(t, rawURL, `<html><body><main><h2>Stealth Render</h2></main></body></html>`), nil
		}

		outputPath := filepath.Join(t.TempDir(), "stealth.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, []string{
			"stealthy-fetch", "https://example.test/protected", outputPath,
			"--css-selector", "h2::text",
			"--block-webrtc",
			"--block-webgl",
			"--hide-canvas",
			"--block-ads",
			"-H", "X-Mode: stealth",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}
		if !seen.Headless || !seen.BlockAds {
			t.Fatalf("stealth browser defaults = %#v", seen)
		}
		if !seen.Stealth.Enabled || !seen.Stealth.GenerateHeaders || !seen.Stealth.BlockWebRTC || !seen.Stealth.DisableWebGL || !seen.Stealth.HideCanvas {
			t.Fatalf("stealth options were not forwarded: %#v", seen.Stealth)
		}
		if got := seen.Headers.Get("X-Mode"); got != "stealth" {
			t.Fatalf("extra header = %q", got)
		}
		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := strings.TrimSpace(string(body)); got != "Stealth Render" {
			t.Fatalf("stealth text = %q", got)
		}
	})
}

func newTestResponse(t *testing.T, rawURL string, body string) *goscrapling.Response {
	t.Helper()
	response, err := goscrapling.NewResponse(strings.NewReader(body), goscrapling.ResponseOptions{
		URL:        rawURL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
	})
	if err != nil {
		t.Fatalf("new test response: %v", err)
	}
	return response
}

func readRequestBody(t *testing.T, r *http.Request) string {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	return string(body)
}
