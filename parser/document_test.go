package parser

import (
	"strings"
	"testing"
)

func TestParseSelectsElementsWithCSS(t *testing.T) {
	html := `<html><body><article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article></body></html>`

	doc, err := Parse(strings.NewReader(html), ParseOptions{URL: "https://example.com/products"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	products := doc.CSS(".product")
	if products.Len() != 1 {
		t.Fatalf("expected 1 product, got %d", products.Len())
	}

	first, ok := products.First()
	if !ok {
		t.Fatal("expected first product")
	}
	if got := first.TagName(); got != "article" {
		t.Fatalf("expected tag article, got %q", got)
	}
	if got := first.Text(); got != "Product 1 Description 1" {
		t.Fatalf("expected normalized text, got %q", got)
	}
	if got, ok := first.Attr("id"); !ok || got != "p1" {
		t.Fatalf("expected id p1, got %q ok=%v", got, ok)
	}
}
