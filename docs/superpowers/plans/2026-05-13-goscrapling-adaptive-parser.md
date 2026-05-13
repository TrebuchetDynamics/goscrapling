# goscrapling Adaptive Parser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first `goscrapling` MVP: parse static HTML, select elements, save adaptive fingerprints, and relocate elements after markup changes.

**Architecture:** Keep the MVP in the root `goscrapling` package with small files for parsing, selections, storage, fingerprinting, scoring, and domain keys. Use `goquery` only for CSS selection and DOM text extraction; keep adaptive logic owned by this project.

**Tech Stack:** Go 1.23 module target, `github.com/PuerkitoBio/goquery@v1.10.3`, Go standard `testing`, strict red-green-refactor.

---

## File Structure

- Create `go.mod`: module metadata for `github.com/TrebuchetDynamics/goscrapling`.
- Create `go.sum`: dependency checksums.
- Create `document.go`: `ParseOptions`, `Document`, `Parse`, `CSS`, `Save`, `Relocate`.
- Create `selection.go`: `Selection` and `Element` wrappers around DOM nodes.
- Create `store.go`: `Key`, `Store`, `MemoryStore`.
- Create `fingerprint.go`: `Fingerprint`, DOM fingerprint extraction, normalization helpers.
- Create `score.go`: deterministic similarity scoring and match threshold.
- Create `domain.go`: adaptive domain derivation from URLs.
- Create `document_test.go`: parser and CSS selection behavior.
- Create `store_test.go`: save, load, keying, and domain isolation behavior.
- Create `relocate_test.go`: adaptive relocation behavior and safeguards.
- Create `example_test.go`: compiling public API example.
- Modify `README.md`: update status after MVP exists.

## Task 1: Initialize Go Module

**Files:**
- Create: `go.mod`
- Create: `go.sum`

- [ ] **Step 1: Initialize the module**

Run:

```bash
go mod init github.com/TrebuchetDynamics/goscrapling
go get github.com/PuerkitoBio/goquery@v1.10.3
```

Expected: `go.mod` exists and pins `github.com/PuerkitoBio/goquery v1.10.3`.

- [ ] **Step 2: Verify empty module metadata**

Run:

```bash
go test ./...
go list -m all
```

Expected: before source files exist, `go test ./...` may print `go: warning: "./..." matched no packages` and `no packages to test`. `go list -m all` must exit successfully and show `github.com/TrebuchetDynamics/goscrapling` plus the pinned dependencies.

- [ ] **Step 3: Commit**

Run:

```bash
git add go.mod go.sum
git commit -m "chore: initialize go module"
```

## Task 2: Parse HTML And Select Elements

**Files:**
- Create: `document_test.go`
- Create: `document.go`
- Create: `selection.go`

- [ ] **Step 1: Write the failing parser test**

Create `document_test.go`:

```go
package goscrapling

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
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./... -run TestParseSelectsElementsWithCSS -count=1
```

Expected: FAIL with undefined `Parse` or `ParseOptions`.

- [ ] **Step 3: Implement minimal parser and selection wrappers**

Create `document.go`:

```go
package goscrapling

import (
	"context"
	"io"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type ParseOptions struct {
	URL   string
	Store Store
}

type Document struct {
	root   *html.Node
	query  *goquery.Document
	url    string
	domain string
	store  Store
}

func Parse(r io.Reader, opts ParseOptions) (*Document, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	query := goquery.NewDocumentFromNode(root)
	return &Document{
		root:   root,
		query:  query,
		url:    opts.URL,
		domain: adaptiveDomain(opts.URL),
		store:  opts.Store,
	}, nil
}

func (d *Document) CSS(selector string) Selection {
	if d == nil || d.query == nil {
		return Selection{}
	}

	var elements []*Element
	d.query.Find(selector).Each(func(_ int, selection *goquery.Selection) {
		for _, node := range selection.Nodes {
			elements = append(elements, &Element{doc: d, node: node})
		}
	})
	return Selection{elements: elements}
}

func (d *Document) Save(ctx context.Context, element *Element, identifier string) error {
	return errAdaptiveNotImplemented
}

func (d *Document) Relocate(ctx context.Context, identifier string) (Match, bool, error) {
	return Match{}, false, errAdaptiveNotImplemented
}
```

Create `selection.go`:

```go
package goscrapling

import (
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type Selection struct {
	elements []*Element
}

func (s Selection) Len() int {
	return len(s.elements)
}

func (s Selection) First() (*Element, bool) {
	if len(s.elements) == 0 {
		return nil, false
	}
	return s.elements[0], true
}

type Element struct {
	doc  *Document
	node *html.Node
}

func (e *Element) TagName() string {
	if e == nil || e.node == nil {
		return ""
	}
	return strings.ToLower(e.node.Data)
}

func (e *Element) Text() string {
	if e == nil || e.node == nil {
		return ""
	}
	return normalizeSpace(goquery.NewDocumentFromNode(e.node).Text())
}

func (e *Element) Attr(name string) (string, bool) {
	if e == nil || e.node == nil {
		return "", false
	}
	name = strings.ToLower(name)
	for _, attr := range e.node.Attr {
		if strings.ToLower(attr.Key) == name {
			return attr.Val, true
		}
	}
	return "", false
}

func normalizeSpace(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}
```

Add a temporary adaptive sentinel to `document.go` or a new small file if the compiler needs it:

```go
package goscrapling

import "errors"

var errAdaptiveNotImplemented = errors.New("adaptive behavior is not implemented")

type Match struct {
	Element *Element
	Score   float64
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
gofmt -w document.go selection.go document_test.go
go test ./... -run TestParseSelectsElementsWithCSS -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add document.go selection.go document_test.go
git commit -m "feat: parse html and select elements"
```

## Task 3: Store Fingerprints By Domain And Identifier

**Files:**
- Create: `store_test.go`
- Create: `store.go`
- Create: `fingerprint.go`
- Create: `domain.go`
- Modify: `document.go`

- [ ] **Step 1: Write the failing storage test**

Create `store_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./... -run TestSaveStoresFingerprintByDomainAndIdentifier -count=1
```

Expected: FAIL with undefined `NewMemoryStore`, `Key`, or storage behavior.

- [ ] **Step 3: Implement keying, memory storage, fingerprints, and Save**

Create `store.go` with `Key`, `Store`, and `MemoryStore`. Create `fingerprint.go` with `Fingerprint` and DOM extraction. Create `domain.go` with `adaptiveDomain`. Update `Document.Save` to validate input, build a key from document domain plus identifier, fingerprint the element, and call `Store.Save`.

Required signatures:

```go
type Key struct {
	Domain     string
	Identifier string
}

type Store interface {
	Save(ctx context.Context, key Key, fp Fingerprint) error
	Load(ctx context.Context, key Key) (Fingerprint, bool, error)
}

func NewMemoryStore() *MemoryStore
```

Required validation errors:

```go
var (
	ErrMissingStore     = errors.New("goscrapling: missing adaptive store")
	ErrNilElement       = errors.New("goscrapling: nil element")
	ErrEmptyIdentifier  = errors.New("goscrapling: empty identifier")
)
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
gofmt -w document.go store.go fingerprint.go domain.go store_test.go
go test ./... -run TestSaveStoresFingerprintByDomainAndIdentifier -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add document.go store.go fingerprint.go domain.go store_test.go
git commit -m "feat: save adaptive fingerprints"
```

## Task 4: Domain Isolation And Default Domain

**Files:**
- Modify: `store_test.go`
- Modify: `domain.go`

- [ ] **Step 1: Write failing domain tests**

Append to `store_test.go`:

```go
func TestSaveIsolatesFingerprintsByDomain(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

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
	store := NewMemoryStore()
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
```

- [ ] **Step 2: Run test to verify it fails if domain handling is incomplete**

Run:

```bash
go test ./... -run 'TestSaveIsolatesFingerprintsByDomain|TestParseUsesDefaultDomainWhenURLIsMissing' -count=1
```

Expected: FAIL if `adaptiveDomain` does not isolate hosts or default missing URLs.

- [ ] **Step 3: Implement domain handling**

Update `adaptiveDomain` so it lowercases the URL hostname, strips a port, and returns `"default"` when URL parsing does not produce a host.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
gofmt -w domain.go store_test.go
go test ./... -run 'TestSaveIsolatesFingerprintsByDomain|TestParseUsesDefaultDomainWhenURLIsMissing' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add domain.go store_test.go
git commit -m "feat: isolate adaptive data by domain"
```

## Task 5: Relocate Elements After Markup Changes

**Files:**
- Create: `relocate_test.go`
- Create: `score.go`
- Modify: `document.go`

- [ ] **Step 1: Write failing relocation tests**

Create `relocate_test.go`:

```go
package goscrapling

import (
	"context"
	"strings"
	"testing"
)

func TestRelocateFindsElementAfterIDMovesToDataID(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
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
	store := NewMemoryStore()
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./... -run 'TestRelocateFindsElementAfterIDMovesToDataID|TestRelocateFindsElementAfterWrapperIsInserted' -count=1
```

Expected: FAIL because `Relocate` still returns the adaptive sentinel error.

- [ ] **Step 3: Implement scoring and Relocate**

Create `score.go` with deterministic scoring helpers. Update `Document.Relocate` to load the stored fingerprint, walk all current document elements, score each candidate, and return the best candidate when its score is at least `0.65`.

Required signatures:

```go
const defaultMinScore = 0.65

type Match struct {
	Element *Element
	Score   float64
}

func scoreFingerprint(candidate Fingerprint, target Fingerprint) float64
```

`scoreFingerprint` must combine tag, text, attribute names, attribute values, parent, sibling tags, and path tags. Ties remain deterministic because candidates are evaluated in document order and the best match only changes when a score is strictly higher.

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
gofmt -w document.go score.go relocate_test.go
go test ./... -run 'TestRelocateFindsElementAfterIDMovesToDataID|TestRelocateFindsElementAfterWrapperIsInserted' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add document.go score.go relocate_test.go
git commit -m "feat: relocate adaptive elements"
```

## Task 6: Matching Safeguards

**Files:**
- Modify: `relocate_test.go`
- Modify: `score.go`

- [ ] **Step 1: Write failing safeguard tests**

Append to `relocate_test.go`:

```go
func TestRelocateReturnsNoMatchBelowThreshold(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
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
	store := NewMemoryStore()
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
	store := NewMemoryStore()
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
```

- [ ] **Step 2: Run tests to verify they fail if scoring is too permissive or nondeterministic**

Run:

```bash
go test ./... -run 'TestRelocateReturnsNoMatchBelowThreshold|TestRelocateChoosesBestMatchingCandidate|TestRelocateTieKeepsDocumentOrder' -count=1
```

Expected: FAIL if the threshold, best-candidate selection, or tie behavior is incomplete.

- [ ] **Step 3: Tighten scoring**

Update `score.go` so unrelated nodes stay below `0.65`, strong matches exceed `0.65`, and ties preserve first candidate by only replacing the best match when `score > best.Score`.

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
gofmt -w score.go relocate_test.go
go test ./... -run 'TestRelocateReturnsNoMatchBelowThreshold|TestRelocateChoosesBestMatchingCandidate|TestRelocateTieKeepsDocumentOrder' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add score.go relocate_test.go
git commit -m "test: cover adaptive matching safeguards"
```

## Task 7: Public Example And README Status

**Files:**
- Create: `example_test.go`
- Modify: `README.md`

- [ ] **Step 1: Write compiling example test**

Create `example_test.go`:

```go
package goscrapling_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
)

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
```

- [ ] **Step 2: Run example to verify it passes**

Run:

```bash
go test ./... -run ExampleDocument_Relocate -count=1
```

Expected: PASS.

- [ ] **Step 3: Update README status**

Change `README.md` current status from planning-only to adaptive parser MVP implementation started, and include the example API shape without claiming browser, crawler, or anti-bot support.

- [ ] **Step 4: Run tests**

Run:

```bash
gofmt -w example_test.go
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add README.md example_test.go
git commit -m "docs: add adaptive parser example"
```

## Task 8: Final Verification

**Files:**
- Modify only if formatting or documentation checks reveal a concrete issue.

- [ ] **Step 1: Run full test suite**

Run:

```bash
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 2: Run formatting and diff checks**

Run:

```bash
gofmt -w *.go
git diff --check
git status --short
```

Expected: no whitespace errors; status only shows intentional files if a final formatting change occurred.

- [ ] **Step 3: Commit final cleanup if needed**

If formatting changed files, run:

```bash
git add .
git commit -m "chore: finalize adaptive parser mvp"
```

If no files changed, do not create an empty commit.
