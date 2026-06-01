package statictools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/fetchers"
)

func TestGormesIntegration(t *testing.T) {
	t.Run("static web_extract returns selector content and structured evidence", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/catalog" {
				t.Fatalf("request path = %q, want /catalog", r.URL.Path)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Catalog</title></head>
<body>
  <nav>Navigation should stay out</nav>
  <article class="product">
    <h2>Alpha Tool</h2>
    <p>Useful static extraction target.</p>
  </article>
  <article class="ad">
    <p>Ad copy should not be returned.</p>
  </article>
</body>
</html>`))
		}))
		defer server.Close()

		adapter := StaticExtractionAdapter{Fetcher: fetchers.Fetcher{Client: server.Client()}}
		response, err := adapter.Call(context.Background(), StaticToolCall{
			Tool:        ToolWebExtract,
			URLs:        []string{server.URL + "/catalog"},
			CSSSelector: "article.product",
		})
		if err != nil {
			t.Fatalf("Call returned error: %v", err)
		}
		if len(response.Results) != 1 {
			t.Fatalf("results len = %d, want one", len(response.Results))
		}

		result := response.Results[0]
		if result.URL != server.URL+"/catalog" || result.Title != "Catalog" {
			t.Fatalf("result metadata = %#v, want final URL and title", result)
		}
		for _, want := range []string{"Alpha Tool", "Useful static extraction target."} {
			if !strings.Contains(result.Content, want) {
				t.Fatalf("selected content missing %q:\n%s", want, result.Content)
			}
		}
		for _, forbidden := range []string{"Navigation should stay out", "Ad copy should not be returned"} {
			if strings.Contains(result.Content, forbidden) {
				t.Fatalf("selected content leaked %q:\n%s", forbidden, result.Content)
			}
		}
		if result.Extraction == nil {
			t.Fatal("expected extraction evidence")
		}
		if got := *result.Extraction; got.Engine != "goscrapling" || got.Mode != "static" || got.StatusCode != http.StatusOK || got.ContentType != "text/html" || got.CSSSelector != "article.product" || got.FinalURL != result.URL {
			t.Fatalf("extraction evidence = %#v, want goscrapling static selector evidence", got)
		}

		payload, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		for _, want := range []string{"\"results\"", "\"extraction\"", "\"css_selector\"", "\"content_type\""} {
			if !strings.Contains(string(payload), want) {
				t.Fatalf("JSON payload %s missing %s", payload, want)
			}
		}
	})
}
