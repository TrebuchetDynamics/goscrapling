package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/core/storage"
)

func TestRelocateFindsElementAfterIDMovesToDataID(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	original, err := Parse(strings.NewReader(`<section><article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article></section>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse original: %v", err)
	}
	element, _ := original.CSS("#p1").First()
	if err := original.Save(ctx, element, "featured-product"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	updated, err := Parse(strings.NewReader(`<section><article class="product new-class" data-id="p1"><div class="info"><h2>Product 1</h2><p>Description 1</p></div></article></section>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse updated: %v", err)
	}

	match, ok, err := updated.Relocate(ctx, "featured-product")
	if err != nil {
		t.Fatalf("Relocate: %v", err)
	}
	if !ok {
		t.Fatal("expected relocated match")
	}
	if got := match.Element.Text(); got != "Product 1 Description 1" {
		t.Fatalf("expected Product 1 text, got %q", got)
	}
	if match.Score < 0.65 {
		t.Fatalf("expected confident score, got %.3f", match.Score)
	}
}

func TestRelocateFindsElementAfterWrapperIsInserted(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	original, err := Parse(strings.NewReader(`<main><section class="products"><article class="product" id="p1">Product 1</article></section></main>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse original: %v", err)
	}
	element, _ := original.CSS("#p1").First()
	if err := original.Save(ctx, element, "featured-product"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	updated, err := Parse(strings.NewReader(`<main><div class="new-container"><section class="products"><article class="product" data-id="p1"><span>Product 1</span></article></section></div></main>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse updated: %v", err)
	}

	match, ok, err := updated.Relocate(ctx, "featured-product")
	if err != nil {
		t.Fatalf("Relocate: %v", err)
	}
	if !ok {
		t.Fatal("expected relocated match")
	}
	if got := match.Element.Text(); got != "Product 1" {
		t.Fatalf("expected Product 1, got %q", got)
	}
}

func TestRelocateReturnsNoMatchBelowThreshold(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	original, err := Parse(strings.NewReader(`<article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse original: %v", err)
	}
	element, _ := original.CSS("#p1").First()
	if err := original.Save(ctx, element, "featured-product"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	updated, err := Parse(strings.NewReader(`<footer><a href="/contact">Contact us</a></footer>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse updated: %v", err)
	}

	_, ok, err := updated.Relocate(ctx, "featured-product")
	if err != nil {
		t.Fatalf("Relocate: %v", err)
	}
	if ok {
		t.Fatal("expected no match for unrelated markup")
	}
}

func TestRelocateChoosesBestMatchingCandidate(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	original, err := Parse(strings.NewReader(`<section><article class="product" id="p1">Product 1</article></section>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse original: %v", err)
	}
	element, _ := original.CSS("#p1").First()
	if err := original.Save(ctx, element, "featured-product"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	updated, err := Parse(strings.NewReader(`<section><article class="product" data-id="p2">Product 2</article><article class="product" data-id="p1">Product 1</article></section>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse updated: %v", err)
	}

	match, ok, err := updated.Relocate(ctx, "featured-product")
	if err != nil {
		t.Fatalf("Relocate: %v", err)
	}
	if !ok {
		t.Fatal("expected match")
	}
	if got := match.Element.Text(); got != "Product 1" {
		t.Fatalf("expected Product 1, got %q", got)
	}
}

func TestRelocateTieKeepsDocumentOrder(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	original, err := Parse(strings.NewReader(`<section><article class="product">Product</article></section>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse original: %v", err)
	}
	element, _ := original.CSS(".product").First()
	if err := original.Save(ctx, element, "product"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	updated, err := Parse(strings.NewReader(`<section><article class="product" data-rank="first">Product</article><article class="product" data-rank="second">Product</article></section>`), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse updated: %v", err)
	}

	match, ok, err := updated.Relocate(ctx, "product")
	if err != nil {
		t.Fatalf("Relocate: %v", err)
	}
	if !ok {
		t.Fatal("expected match")
	}
	if got, _ := match.Element.Attr("data-rank"); got != "first" {
		t.Fatalf("expected first document-order match, got %q", got)
	}
}
