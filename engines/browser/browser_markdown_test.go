package browser

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestBrowserFetchMarkdownDump(t *testing.T) {
	engine := &recordingBrowserEngine{
		result: BrowserResult{
			URL:        "https://example.com/rendered",
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			Body: []byte(`<!doctype html>
<html>
<head>
  <title>ignored title</title>
  <style>.hidden { display: none }</style>
  <script>window.secret = "do not keep"</script>
</head>
<body>
  <nav><a href="/menu">Menu Link</a></nav>
  <main>
    <article>
      <h1>Rendered Browser Title</h1>
      <p>Use <a href="/docs">documentation</a> for examples.</p>
      <h2>Details</h2>
      <p>JavaScript rendered article text.</p>
    </article>
  </main>
</body>
</html>`),
		},
	}
	fetcher := BrowserFetcher{Engine: engine}

	dump, err := fetcher.FetchMarkdown(context.Background(), "https://example.com/app", BrowserOptions{
		Headers: http.Header{"X-Test": []string{"markdown"}},
		LoadDOM: true,
	})
	if err != nil {
		t.Fatalf("FetchMarkdown returned error: %v", err)
	}

	if engine.request.URL != "https://example.com/app" || !engine.request.LoadDOM {
		t.Fatalf("expected FetchMarkdown to use BrowserFetcher options, got %#v", engine.request)
	}
	if got := engine.request.Headers.Get("X-Test"); got != "markdown" {
		t.Fatalf("request header = %q, want markdown", got)
	}

	wantParts := []string{
		"# Rendered Browser Title",
		"Use [documentation](/docs) for examples.",
		"## Details",
		"JavaScript rendered article text.",
	}
	for _, want := range wantParts {
		if !strings.Contains(dump.Markdown, want) {
			t.Fatalf("markdown missing %q in:\n%s", want, dump.Markdown)
		}
	}

	for _, unwanted := range []string{"Menu Link", "do not keep", ".hidden", "<script", "<style"} {
		if strings.Contains(dump.Markdown, unwanted) {
			t.Fatalf("markdown kept unwanted %q in:\n%s", unwanted, dump.Markdown)
		}
	}
	if dump.URL != "https://example.com/rendered" || dump.StatusCode != http.StatusOK {
		t.Fatalf("dump metadata = url %q status %d", dump.URL, dump.StatusCode)
	}
}
