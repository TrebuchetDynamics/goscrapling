package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/core/storage"
)

func TestSaveStoresFingerprintByDomainAndIdentifier(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	doc, err := Parse(strings.NewReader(`<article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article>`), ParseOptions{
		URL:   "https://example.com/products",
		Store: store,
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	element, ok := doc.CSS(".product").First()
	if !ok {
		t.Fatal("expected product")
	}

	if err := doc.Save(ctx, element, "featured-product"); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	fp, ok, err := store.Load(ctx, Key{Domain: "example.com", Identifier: "featured-product"})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected stored fingerprint")
	}
	if fp.Tag != "article" {
		t.Fatalf("expected article fingerprint, got %q", fp.Tag)
	}
	if fp.Text != "Product 1 Description 1" {
		t.Fatalf("expected normalized text, got %q", fp.Text)
	}
	if fp.Attributes["id"] != "p1" {
		t.Fatalf("expected id p1, got %q", fp.Attributes["id"])
	}
}

func TestSaveIsolatesFingerprintsByDomain(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()

	first, err := Parse(strings.NewReader(`<article class="product" id="p1">First</article>`), ParseOptions{URL: "https://example.com/a", Store: store})
	if err != nil {
		t.Fatalf("Parse first: %v", err)
	}
	second, err := Parse(strings.NewReader(`<article class="product" id="p2">Second</article>`), ParseOptions{URL: "https://other.example/a", Store: store})
	if err != nil {
		t.Fatalf("Parse second: %v", err)
	}

	firstElement, _ := first.CSS(".product").First()
	secondElement, _ := second.CSS(".product").First()
	if err := first.Save(ctx, firstElement, "shared"); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	if err := second.Save(ctx, secondElement, "shared"); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	firstFP, ok, err := store.Load(ctx, Key{Domain: "example.com", Identifier: "shared"})
	if err != nil || !ok {
		t.Fatalf("Load first ok=%v err=%v", ok, err)
	}
	secondFP, ok, err := store.Load(ctx, Key{Domain: "other.example", Identifier: "shared"})
	if err != nil || !ok {
		t.Fatalf("Load second ok=%v err=%v", ok, err)
	}
	if firstFP.Text != "First" || secondFP.Text != "Second" {
		t.Fatalf("expected isolated fingerprints, got %q and %q", firstFP.Text, secondFP.Text)
	}
}

func TestParseUsesDefaultDomainWhenURLIsMissing(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	doc, err := Parse(strings.NewReader(`<article class="product">Default</article>`), ParseOptions{Store: store})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	element, _ := doc.CSS(".product").First()
	if err := doc.Save(ctx, element, "item"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	fp, ok, err := store.Load(ctx, Key{Domain: "default", Identifier: "item"})
	if err != nil || !ok {
		t.Fatalf("Load ok=%v err=%v", ok, err)
	}
	if fp.Text != "Default" {
		t.Fatalf("expected Default text, got %q", fp.Text)
	}
}
