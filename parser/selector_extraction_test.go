package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestSelectorExtractionHelpers(t *testing.T) {
	body := `
		<html>
			<body>
				<article class="product" data-sku="sku-1" data-kind="featured">
					<h1>Trail <span>Kit</span></h1>
					<a href="/products/trail-kit">Details</a>
					<span class="price">$51.77</span>
					<script type="application/json" id="payload">{"name":"Trail Kit","count":2}</script>
				</article>
				<article class="product" data-sku="sku-2">
					<h1>City Pack</h1>
					<a href="/products/city-pack">Details</a>
					<span class="price">$42.10</span>
				</article>
			</body>
		</html>`

	doc, err := Parse(strings.NewReader(body), ParseOptions{URL: "https://example.com/products"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	t.Run("text pseudo element extracts direct text nodes", func(t *testing.T) {
		titles := doc.CSS(".product h1::text")

		if got := titles.Get(); got != "Trail" {
			t.Fatalf("expected first title direct text Trail, got %q", got)
		}
		if got := titles.GetAll().Strings(); !equalStrings(got, []string{"Trail", "City Pack"}) {
			t.Fatalf("expected all title direct text nodes, got %#v", got)
		}
		if got := titles.Text(); got != "Trail\nCity Pack" {
			t.Fatalf("expected selection text from extracted values, got %q", got)
		}
	})

	t.Run("attr pseudo element extracts attributes", func(t *testing.T) {
		links := doc.CSS(".product a::attr(href)")

		if got := links.Get(); got != "/products/trail-kit" {
			t.Fatalf("expected first href, got %q", got)
		}
		if got := links.GetAll().Strings(); !equalStrings(got, []string{"/products/trail-kit", "/products/city-pack"}) {
			t.Fatalf("expected href values, got %#v", got)
		}
	})

	t.Run("regex extraction works on selected text and attributes", func(t *testing.T) {
		prices, err := doc.CSS(".price::text").Regex(`\d+\.\d{2}`)
		if err != nil {
			t.Fatalf("Regex: %v", err)
		}
		if got := prices.Strings(); !equalStrings(got, []string{"51.77", "42.10"}) {
			t.Fatalf("expected price regex matches, got %#v", got)
		}

		slug, err := doc.CSS(".product a::attr(href)").RegexFirst(`/products/([^/]+)`)
		if err != nil {
			t.Fatalf("RegexFirst: %v", err)
		}
		if slug != "trail-kit" {
			t.Fatalf("expected first product slug, got %q", slug)
		}
	})

	t.Run("json parsing works from extracted text", func(t *testing.T) {
		payload, err := doc.CSS("#payload::text").JSON()
		if err != nil {
			t.Fatalf("JSON: %v", err)
		}

		object, ok := payload.(map[string]any)
		if !ok {
			t.Fatalf("expected JSON object, got %T", payload)
		}
		if object["name"] != "Trail Kit" || object["count"] != float64(2) {
			t.Fatalf("unexpected JSON payload: %#v", object)
		}
	})

	t.Run("text handler and text handlers expose typed helpers", func(t *testing.T) {
		text := TextHandler("  Trail\n\tKit  ")
		if got := text.Clean(); got != "Trail Kit" {
			t.Fatalf("expected cleaned text, got %q", got)
		}

		handlers := TextHandlers{TextHandler("sku-1"), TextHandler("sku-2")}
		if got := handlers.Get(); got != "sku-1" {
			t.Fatalf("expected first text handler, got %q", got)
		}
		matches, err := handlers.Regex(`sku-(\d)`)
		if err != nil {
			t.Fatalf("TextHandlers.Regex: %v", err)
		}
		if got := matches.Strings(); !equalStrings(got, []string{"1", "2"}) {
			t.Fatalf("expected flattened capture groups, got %#v", got)
		}
	})

	t.Run("attributes handler exposes element attributes as text handlers", func(t *testing.T) {
		product, ok := doc.CSS(".product").First()
		if !ok {
			t.Fatal("expected first product")
		}

		attrs := product.Attrs()
		if attrs.Len() != 3 {
			t.Fatalf("expected three attributes, got %d", attrs.Len())
		}
		sku, ok := attrs.Get("data-sku")
		if !ok || sku != "sku-1" {
			t.Fatalf("expected data-sku sku-1, got %q ok=%v", sku, ok)
		}
		matches := attrs.SearchValues("featured", false)
		if len(matches) != 1 {
			t.Fatalf("expected one exact attribute value match, got %d", len(matches))
		}
		kind, ok := matches[0].Get("data-kind")
		if !ok || kind != "featured" {
			t.Fatalf("expected matched data-kind featured, got %q ok=%v", kind, ok)
		}
	})

	t.Run("invalid pseudo element selector returns a predictable error", func(t *testing.T) {
		_, err := doc.SelectCSS(nil, "a::attr()", SelectorOptions{})
		if !errors.Is(err, ErrInvalidSelector) {
			t.Fatalf("expected ErrInvalidSelector, got %v", err)
		}
	})
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
