package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestSQLiteStore(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "adaptive-store.sqlite")

	store, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStore returned error: %v", err)
	}

	first := Fingerprint{
		Tag:              "article",
		Text:             "First product",
		Attributes:       map[string]string{"id": "p1", "class": "product"},
		ParentTag:        "section",
		ParentAttributes: map[string]string{"id": "catalog"},
		SiblingTags:      []string{"article", "aside"},
		ChildrenTags:     []string{"h2", "p"},
		PathTags:         []string{"html", "body", "section", "article"},
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
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("reopen SQLite store: %v", err)
	}
	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Errorf("Close reopened store: %v", err)
		}
	})

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

	if err := reopened.Close(); err != nil {
		t.Fatalf("Close reopened before closed behavior check: %v", err)
	}
	if err := reopened.Save(ctx, Key{Domain: "example.com", Identifier: "closed"}, first); !errors.Is(err, ErrClosedStore) {
		t.Fatalf("Save after Close error = %v, want ErrClosedStore", err)
	}
	if _, _, err := reopened.Load(ctx, Key{Domain: "example.com", Identifier: "featured"}); !errors.Is(err, ErrClosedStore) {
		t.Fatalf("Load after Close error = %v, want ErrClosedStore", err)
	}

	incompatiblePath := filepath.Join(t.TempDir(), "incompatible.sqlite")
	if err := writeSQLiteUserVersion(incompatiblePath, 999); err != nil {
		t.Fatalf("write incompatible SQLite store: %v", err)
	}
	_, err = NewSQLiteStore(incompatiblePath)
	if !errors.Is(err, ErrUnsupportedStoreSchema) {
		t.Fatalf("expected unsupported schema error, got %v", err)
	}
}

func writeSQLiteUserVersion(path string, version int) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("PRAGMA user_version = " + sqlInt(version))
	return err
}

func sqlInt(value int) string {
	return strconv.Itoa(value)
}
