package goscrapling_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
)

func TestExamples(t *testing.T) {
	t.Run("public APIs support README examples", func(t *testing.T) {
		doc, err := goscrapling.Parse(strings.NewReader(`<main><article class="product" data-sku="A-42">Trail pack</article></main>`), goscrapling.ParseOptions{URL: "https://example.com/products"})
		if err != nil {
			t.Fatalf("parse example HTML: %v", err)
		}
		if got := doc.CSS(".product::text").Get().String(); got != "Trail pack" {
			t.Fatalf("CSS text = %q, want Trail pack", got)
		}
		if got := doc.CSS(".product::attr(data-sku)").Get().String(); got != "A-42" {
			t.Fatalf("CSS attr = %q, want A-42", got)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><body><h1>Local Product</h1></body></html>`))
		}))
		defer server.Close()

		fetcher := goscrapling.Fetcher{Client: server.Client()}
		response, err := fetcher.Get(server.URL, goscrapling.RequestOptions{})
		if err != nil {
			t.Fatalf("fetch local example page: %v", err)
		}
		if got := response.CSS("h1::text").Get().String(); got != "Local Product" {
			t.Fatalf("fetched title = %q, want Local Product", got)
		}
	})

	t.Run("README maps upstream feature groups to Go status", func(t *testing.T) {
		body, err := os.ReadFile("README.md")
		if err != nil {
			t.Fatalf("read README.md: %v", err)
		}
		readme := string(body)
		for _, phrase := range []string{
			"## Parity Status Map",
			"Parser and selectors",
			"Adaptive relocation",
			"Static fetcher and response",
			"Browser fetching",
			"Spider runtime",
			"development response cache",
			"scripted `shell -c`",
			"MCP and Gormes integration",
			"static Gormes `web_extract`",
			"Migration guidance",
			"Translated README files, assets, stylesheets, and ReadTheDocs config",
		} {
			if !strings.Contains(readme, phrase) {
				t.Fatalf("README.md missing public parity phrase %q", phrase)
			}
		}
	})
}

func ExampleDocument_Relocate() {
	ctx := context.Background()
	store := goscrapling.NewMemoryStore()

	before, _ := goscrapling.Parse(strings.NewReader(`<article class="product" id="p1">Product 1</article>`), goscrapling.ParseOptions{
		URL:   "https://example.com/products",
		Store: store,
	})
	element, _ := before.CSS("#p1").First()
	_ = before.Save(ctx, element, "featured-product")

	after, _ := goscrapling.Parse(strings.NewReader(`<article class="product" data-id="p1"><span>Product 1</span></article>`), goscrapling.ParseOptions{
		URL:   "https://example.com/products",
		Store: store,
	})
	match, ok, _ := after.Relocate(ctx, "featured-product")
	fmt.Println(ok, match.Element.Text())

	// Output:
	// true Product 1
}
