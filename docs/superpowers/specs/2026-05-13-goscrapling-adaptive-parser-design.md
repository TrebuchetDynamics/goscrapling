# goscrapling Adaptive Parser MVP Design

Date: 2026-05-13

Status: approved direction, pending written-spec review.

## Summary

`goscrapling` will start as a Go-native adaptive parser library inspired by D4Vinci/Scrapling. The first milestone is not a crawler, browser automation layer, or anti-bot system. It is the smallest useful core: parse static HTML, select elements, save stable fingerprints, and relocate the same logical element after markup changes.

This design avoids claiming official port status. Public framing is: "Go-native adaptive web scraping inspired by D4Vinci/Scrapling."

## Goals

- Provide a Go API for parsing HTML and selecting elements.
- Persist an element fingerprint under a domain plus identifier.
- Relocate an element when the original selector fails because markup changed.
- Make the adaptive matching algorithm deterministic and testable.
- Build the first implementation with strict TDD.
- Keep future fetcher, browser, crawler, CLI, and Gormes integration paths open.

## Non-Goals

- No browser automation in the first milestone.
- No anti-bot bypass claims in the first milestone.
- No full Scrapling API compatibility promise.
- No crawler or spider engine in the first milestone.
- No MCP server or Gormes tool in the first milestone.
- No copying upstream Scrapling source code.

## Users

Primary users are Go developers who need resilient selectors for scraping workflows.

Secondary users are OpenClaw/Gormes agents that may later call `goscrapling` as a web search or extraction tool.

## Package Shape

The initial public package can stay small:

- `Document`: parsed HTML document plus page metadata.
- `Element`: wrapper around one DOM element.
- `Selection`: ordered collection of elements.
- `Fingerprint`: serializable adaptive element identity.
- `Store`: interface for adaptive persistence.
- `MemoryStore`: test and simple runtime store.
- `Relocator`: scores candidates and returns matches.

Suggested future package split:

- root package: parser, selector, adaptive API.
- `storage`: durable stores such as SQLite.
- `fetch`: HTTP and browser-backed fetchers.
- `crawl`: spider engine, scheduler, checkpointing.
- `cmd/goscrapling`: CLI.
- `tools/gormes`: optional Gormes/OpenClaw integration.

For phase 1, keep files in the root package unless the code naturally demands a split.

## Public API Sketch

The exact API will be finalized through tests, but the intended shape is:

```go
doc, err := goscrapling.Parse(strings.NewReader(html), goscrapling.ParseOptions{
    URL:   "https://example.com/products",
    Store: store,
})

products := doc.CSS(".product")
err = doc.Save(products.First(), "featured-product")

updated, err := goscrapling.Parse(strings.NewReader(changedHTML), goscrapling.ParseOptions{
    URL:   "https://example.com/products",
    Store: store,
})

match, ok, err := updated.Relocate("featured-product")
```

The first implementation should optimize for clarity over convenience. Fluent helpers can be added only after core behavior is proven.

## Adaptive Key Model

Adaptive data is keyed by:

- Domain: derived from `ParseOptions.URL`, with a stable `"default"` fallback when absent.
- Identifier: explicit caller-provided string, or later the selector string for helper methods.

Phase 1 requires explicit identifiers. Selector-based `auto_save` helpers can come after the core relocate behavior is stable.

## Fingerprint Model

The first `Fingerprint` should include:

- Element tag name.
- Normalized element text.
- Attribute names and values.
- Parent tag name.
- Parent normalized text.
- Parent attribute names and values.
- Sibling tag names.
- Path tag names from document root to element.

Normalization rules:

- Trim leading and trailing whitespace.
- Collapse internal whitespace.
- Lowercase tag and attribute names.
- Preserve attribute values except whitespace normalization.
- Keep class tokens as values, but compare them token-aware.

## Matching Algorithm

`Relocator` scores candidate elements from the current document against the stored fingerprint.

Initial scoring weights:

- Tag name match: high weight.
- Text similarity: high weight when non-empty.
- Attribute name overlap: medium weight.
- Attribute value similarity: medium weight.
- Parent tag and parent attributes: medium weight.
- Sibling tag similarity: low to medium weight.
- Path tag similarity: low to medium weight.

The score must be deterministic. Ties should resolve by document order.

The algorithm should return:

- Best element.
- Boolean found flag.
- Score value for diagnostics.
- Optional candidate scores later for debugging.

Minimum threshold should be conservative enough to avoid returning unrelated nodes. The threshold should be driven by tests rather than hard-coded intuition.

## Storage

Use an interface first:

```go
type Store interface {
    Save(ctx context.Context, key Key, fp Fingerprint) error
    Load(ctx context.Context, key Key) (Fingerprint, bool, error)
}
```

`MemoryStore` is required for phase 1. A durable SQLite store is phase 2 or later because it adds dependency and migration concerns that are not needed to prove adaptive relocation.

## Error Handling

Errors should be explicit and boring:

- Parse failures return errors from `Parse`.
- Saving a nil or absent element returns a typed error.
- Relocating with missing store data returns `ok=false, err=nil`.
- Storage failures return `err`.
- Invalid URLs do not fail parsing; they fall back to the default domain unless a strict option is added later.

## Testing Strategy

Implementation must follow red-green-refactor.

The first test sequence should cover:

1. Parse simple HTML and select by CSS.
2. Save a selected element fingerprint into `MemoryStore`.
3. Load the fingerprint by domain plus identifier.
4. Relocate an element after `id` changes to `data-id`.
5. Relocate an element after wrapper markup is inserted.
6. Prefer the better match when two candidates share the same tag.
7. Return no match below the threshold.
8. Keep domain isolation between two URLs.
9. Resolve ties by document order.
10. Preserve deterministic results across repeated runs.

No production code should be written before the corresponding failing test is observed.

## Documentation Strategy

Documentation comes before implementation:

- Research map: `docs/research/scrapling-architecture-map.md`
- Go OSS survey: `docs/research/go-scraping-oss-survey.md`
- Design spec: this file
- Later implementation plan: written only after this spec is reviewed

The public README must state that the project is inspired by Scrapling and not affiliated with it.

## Future Phases

Phase 2: durable storage and selector convenience.

- SQLite store.
- `CSS(..., AutoSave(...))` or a simple equivalent.
- `CSSAdaptive(...)` helper if tests prove the API is clear.

Phase 3: fetchers.

- Static HTTP fetcher.
- Response type that embeds document behavior and HTTP metadata.
- Dependency review for advanced HTTP impersonation.

Phase 4: browser-backed fetching.

- Compare `rod` and `chromedp`.
- Implement browser fetcher behind a small interface.
- Add network-idle and wait-selector behavior if tests can cover it.

Phase 5: crawling.

- Scheduler.
- Request fingerprints.
- Concurrency and per-domain limits.
- Cache and checkpoint support.

Phase 6: tool integration.

- CLI.
- Gormes/OpenClaw web search adapter.
- MCP server only if there is a concrete consumer.

## Acceptance Criteria For Phase 1

- `go test ./...` passes.
- Adaptive relocation is covered by static HTML fixtures.
- Public API examples compile.
- README and design docs accurately describe current status.
- No browser, crawler, or anti-bot claims are made before implementation exists.

## Open Decisions Resolved For This Spec

- Start with adaptive parser MVP instead of crawler or browser layer.
- Keep upstream Scrapling as local reference material only.
- Prefer explicit identifiers in phase 1.
- Use `MemoryStore` first.
- Defer durable SQLite, browser fetchers, crawler, CLI, MCP, and Gormes integration.

