package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/core/storage"
)

func TestAdaptiveSelectorModes(t *testing.T) {
	t.Run("auto save uses selector as default key", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		doc := mustParseAdaptiveDocument(t, `<article class="product" id="p1">Product 1</article>`, "https://example.com/products", store)

		selection, err := doc.SelectCSS(ctx, ".product", SelectorOptions{AutoSave: true})
		if err != nil {
			t.Fatalf("SelectCSS: %v", err)
		}
		if selection.Len() != 1 {
			t.Fatalf("expected selector result, got %d", selection.Len())
		}

		if _, ok, err := store.Load(ctx, Key{Domain: "example.com", Identifier: ".product"}); err != nil || !ok {
			t.Fatalf("expected selector key save, ok=%v err=%v", ok, err)
		}
	})

	t.Run("adaptive fallback returns relocated element on miss", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDocument(t, `<section><article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article></section>`, "https://example.com/products", store)
		if _, err := before.SelectCSS(ctx, "#p1", SelectorOptions{AutoSave: true}); err != nil {
			t.Fatalf("auto save: %v", err)
		}

		after := mustParseAdaptiveDocument(t, `<section><article class="product new-class" data-id="p1"><div><h2>Product 1</h2><p>Description 1</p></div></article></section>`, "https://example.com/products", store)
		selection, err := after.SelectCSS(ctx, "#p1", SelectorOptions{Adaptive: true})
		if err != nil {
			t.Fatalf("adaptive select: %v", err)
		}
		if selection.Len() != 1 {
			t.Fatalf("expected one relocated element, got %d", selection.Len())
		}
		element, _ := selection.First()
		if got, ok := element.Attr("data-id"); !ok || got != "p1" {
			t.Fatalf("expected relocated data-id p1, got %q ok=%v", got, ok)
		}
	})

	t.Run("explicit identifier overrides selector key", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDocument(t, `<article class="product" id="p1">Product 1</article>`, "https://example.com/products", store)
		if _, err := before.SelectCSS(ctx, ".product", SelectorOptions{Identifier: "featured", AutoSave: true}); err != nil {
			t.Fatalf("auto save: %v", err)
		}

		after := mustParseAdaptiveDocument(t, `<article class="product" data-id="p1">Product 1</article>`, "https://example.com/products", store)
		bySelector, err := after.SelectCSS(ctx, "#missing", SelectorOptions{Adaptive: true})
		if err != nil {
			t.Fatalf("selector key adaptive select: %v", err)
		}
		if bySelector.Len() != 0 {
			t.Fatalf("expected selector key miss, got %d", bySelector.Len())
		}

		byIdentifier, err := after.SelectCSS(ctx, "#missing", SelectorOptions{Identifier: "featured", Adaptive: true})
		if err != nil {
			t.Fatalf("identifier adaptive select: %v", err)
		}
		if byIdentifier.Len() != 1 {
			t.Fatalf("expected identifier relocation, got %d", byIdentifier.Len())
		}
	})

	t.Run("combined selectors save separate default keys", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDocument(t, `<section><article id="p1">Product 1</article><article id="p2">Product 2</article></section>`, "https://example.com/products", store)
		selection, err := before.SelectCSS(ctx, "#p1, #p2", SelectorOptions{AutoSave: true})
		if err != nil {
			t.Fatalf("auto save combined selectors: %v", err)
		}
		if selection.Len() != 2 {
			t.Fatalf("expected two selector results, got %d", selection.Len())
		}

		after := mustParseAdaptiveDocument(t, `<section><article data-id="p1">Product 1</article><article data-id="p2">Product 2</article></section>`, "https://example.com/products", store)
		relocated, err := after.SelectCSS(ctx, "#p2", SelectorOptions{Adaptive: true})
		if err != nil {
			t.Fatalf("adaptive select p2: %v", err)
		}
		element, ok := relocated.First()
		if !ok {
			t.Fatal("expected relocated p2")
		}
		if got, ok := element.Attr("data-id"); !ok || got != "p2" {
			t.Fatalf("expected relocated p2, got %q ok=%v", got, ok)
		}
	})

	t.Run("percentage controls relocation threshold", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDocument(t, `<article class="product" id="p1">Product 1</article>`, "https://example.com/products", store)
		if _, err := before.SelectCSS(ctx, "#p1", SelectorOptions{AutoSave: true}); err != nil {
			t.Fatalf("auto save: %v", err)
		}

		after := mustParseAdaptiveDocument(t, `<article class="product" data-id="p1"><span>Product 1</span></article>`, "https://example.com/products", store)
		low, err := after.SelectCSS(ctx, "#p1", SelectorOptions{Adaptive: true, Percentage: 40})
		if err != nil {
			t.Fatalf("low threshold adaptive select: %v", err)
		}
		if low.Len() != 1 {
			t.Fatalf("expected low threshold relocation, got %d", low.Len())
		}

		high, err := after.SelectCSS(ctx, "#p1", SelectorOptions{Adaptive: true, Percentage: 99})
		if err != nil {
			t.Fatalf("high threshold adaptive select: %v", err)
		}
		if high.Len() != 0 {
			t.Fatalf("expected high threshold miss, got %d", high.Len())
		}
	})

	t.Run("domain override changes storage key", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDocument(t, `<article class="product" id="p1">Product 1</article>`, "https://first.example/products", store)
		if _, err := before.SelectCSS(ctx, "#p1", SelectorOptions{AutoSave: true, Domain: "shared.example"}); err != nil {
			t.Fatalf("auto save with domain override: %v", err)
		}

		after := mustParseAdaptiveDocument(t, `<article class="product" data-id="p1">Product 1</article>`, "https://second.example/products", store)
		withoutOverride, err := after.SelectCSS(ctx, "#p1", SelectorOptions{Adaptive: true})
		if err != nil {
			t.Fatalf("adaptive select without override: %v", err)
		}
		if withoutOverride.Len() != 0 {
			t.Fatalf("expected domain-isolated miss, got %d", withoutOverride.Len())
		}

		withOverride, err := after.SelectCSS(ctx, "#p1", SelectorOptions{Adaptive: true, Domain: "shared.example"})
		if err != nil {
			t.Fatalf("adaptive select with override: %v", err)
		}
		if withOverride.Len() != 1 {
			t.Fatalf("expected domain override relocation, got %d", withOverride.Len())
		}
	})

	t.Run("retrieve wraps relocation", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDocument(t, `<article class="product" id="p1">Product 1</article>`, "https://example.com/products", store)
		element, _ := before.CSS("#p1").First()
		if err := before.Save(ctx, element, "featured"); err != nil {
			t.Fatalf("Save: %v", err)
		}

		after := mustParseAdaptiveDocument(t, `<article class="product" data-id="p1">Product 1</article>`, "https://example.com/products", store)
		relocated, ok, err := after.Retrieve(ctx, "featured")
		if err != nil {
			t.Fatalf("Retrieve: %v", err)
		}
		if !ok {
			t.Fatal("expected retrieved relocation")
		}
		if got, ok := relocated.Attr("data-id"); !ok || got != "p1" {
			t.Fatalf("expected retrieved data-id p1, got %q ok=%v", got, ok)
		}
	})
}

func mustParseAdaptiveDocument(t *testing.T, body string, rawURL string, store Store) *Document {
	t.Helper()
	doc, err := Parse(strings.NewReader(body), ParseOptions{URL: rawURL, Store: store})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}
