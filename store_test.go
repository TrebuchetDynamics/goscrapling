package goscrapling

import (
	"context"
	"strings"
	"testing"
)

func TestSaveStoresFingerprintByDomainAndIdentifier(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
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
