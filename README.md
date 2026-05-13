# goscrapling

Go-native Scrapling-style web scraping.

`goscrapling` is a long-term Go-native feature port of D4Vinci/Scrapling. The goal is Scrapling-style parity across parser APIs, adaptive element relocation, fetchers, browser-backed fetching, spiders, CLI workflows, and agent/tool integration.

This project is not affiliated with D4Vinci/Scrapling. It is an independent Go implementation that uses Scrapling as public reference material.

## Current Status

Very early and far from Scrapling parity.

Implemented now:

- Parse static HTML into queryable selectors.
- Select elements with CSS-like APIs.
- Save an element fingerprint under a domain plus identifier.
- Relocate the same logical element after markup changes.
- Drive all behavior with test-first fixtures.

Missing major Scrapling subsystems:

- HTTP fetchers.
- Browser-backed fetching.
- Response object with HTTP metadata.
- Fetcher sessions and request option merging.
- Proxy rotation.
- Crawling and spider scheduling.
- Robots, cache, checkpoints, and stats.
- CLI, MCP, or Gormes/OpenClaw tool integration.

The current adaptive parser is phase 0. It is not enough for the project to be considered a real Scrapling-style port.

## Port Target

The parity target is documented in:

- [Scrapling Parity Matrix](docs/research/scrapling-parity-matrix.md)
- [True Port Design](docs/superpowers/specs/2026-05-13-goscrapling-true-port-design.md)
- [Scrapling Feature Map](docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md)
- [Upstream Coverage Ledger](docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md)
- [Progress Ledger](docs/content/building-goscrapling/architecture_plan/progress.json)

Next required milestone:

1. Add a `Response` type that behaves like a parsed document plus HTTP metadata.
2. Add a static `Fetcher` and `FetcherSession`.
3. Test fetch behavior against local HTTP fixtures.
4. Update the parity matrix as each Scrapling subsystem moves from planned to partial or done.

## Example

```go
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
```

## Reference Material

The upstream Scrapling repository is cloned locally for study at:

`references/Scrapling`

The local clone is ignored by git. Public documentation records only the observed architecture and decisions, not copied source.

## Documentation

- [Scrapling Architecture Map](docs/research/scrapling-architecture-map.md)
- [Scrapling Parity Matrix](docs/research/scrapling-parity-matrix.md)
- [Scrapling Feature Map](docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md)
- [Upstream Coverage Ledger](docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md)
- [Progress Schema](docs/content/building-goscrapling/builder-loop/progress-schema.md)
- [Go Scraping OSS Survey](docs/research/go-scraping-oss-survey.md)
- [Adaptive Parser MVP Design](docs/superpowers/specs/2026-05-13-goscrapling-adaptive-parser-design.md)
- [True Port Design](docs/superpowers/specs/2026-05-13-goscrapling-true-port-design.md)
