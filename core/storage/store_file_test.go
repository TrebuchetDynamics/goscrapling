package storage

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileStore(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "adaptive-store.json")

	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore returned error: %v", err)
	}

	first := Fingerprint{
		Tag:              "article",
		Text:             "First product",
		Attributes:       map[string]string{"id": "p1"},
		ParentAttributes: map[string]string{},
		PathTags:         []string{"html", "body", "article"},
	}
	second := Fingerprint{
		Tag:              "article",
		Text:             "Other domain product",
		Attributes:       map[string]string{"id": "p2"},
		ParentAttributes: map[string]string{},
		PathTags:         []string{"html", "body", "article"},
	}

	if err := store.Save(ctx, Key{Domain: "example.com", Identifier: "featured"}, first); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	if err := store.Save(ctx, Key{Domain: "other.example", Identifier: "featured"}, second); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	reopened, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("reopen file store: %v", err)
	}

	gotFirst, ok, err := reopened.Load(ctx, Key{Domain: "example.com", Identifier: "featured"})
	if err != nil || !ok {
		t.Fatalf("Load first ok=%v err=%v", ok, err)
	}
	gotSecond, ok, err := reopened.Load(ctx, Key{Domain: "other.example", Identifier: "featured"})
	if err != nil || !ok {
		t.Fatalf("Load second ok=%v err=%v", ok, err)
	}

	if !reflect.DeepEqual(gotFirst, first) {
		t.Fatalf("first fingerprint mismatch:\nwant: %#v\n got: %#v", first, gotFirst)
	}
	if !reflect.DeepEqual(gotSecond, second) {
		t.Fatalf("second fingerprint mismatch:\nwant: %#v\n got: %#v", second, gotSecond)
	}

	gotFirst.Attributes["id"] = "changed"
	gotFirst.PathTags[0] = "changed"
	again, ok, err := reopened.Load(ctx, Key{Domain: "example.com", Identifier: "featured"})
	if err != nil || !ok {
		t.Fatalf("Load first again ok=%v err=%v", ok, err)
	}
	if again.Attributes["id"] != "p1" || again.PathTags[0] != "html" {
		t.Fatalf("expected loaded fingerprints to be copied, got %#v", again)
	}

	missing, ok, err := reopened.Load(ctx, Key{Domain: "missing.example", Identifier: "featured"})
	if err != nil {
		t.Fatalf("Load missing returned error: %v", err)
	}
	if ok || !reflect.DeepEqual(missing, Fingerprint{}) {
		t.Fatalf("expected missing fingerprint to be absent, got ok=%v fp=%#v", ok, missing)
	}

	_, err = NewFileStore(filepath.Join("..", "..", "testdata", "adaptive_store", "incompatible-schema.json"))
	if !errors.Is(err, ErrUnsupportedStoreSchema) {
		t.Fatalf("expected unsupported schema error, got %v", err)
	}
}
