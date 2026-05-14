# goscrapling

Go-native Scrapling-style web scraping.

`goscrapling` is a long-term Go-native feature port of D4Vinci/Scrapling. The goal is Scrapling-style parity across parser APIs, adaptive element relocation, fetchers, browser-backed fetching, spiders, CLI workflows, and agent/tool integration.

This project is not affiliated with D4Vinci/Scrapling. It is an independent Go implementation that uses Scrapling as public reference material.

## Why This Project Exists

goscrapling is useful if it stays honest about scope: a Go-native extraction
engine moving toward Scrapling-style behavior with tests and progress rows, not
a claimed complete clone.

The portfolio value is the engineering method: upstream study, parity ledgers,
small tested slices, and a realistic downstream integration target. The Gormes
value is a single-binary-friendly extraction core that can eventually power
agent web tools without a Python sidecar.

Recommended positioning:

> goscrapling is a Go-native web extraction engine inspired by Scrapling, built
> for agent runtimes and single-binary deployments.

See [Portfolio and Gormes Fit](docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md)
for the product boundary and integration rationale.

## Current Status

Very early and far from Scrapling parity.

Implemented now:

- Parse static HTML into queryable selectors.
- Select elements with CSS-like APIs.
- Save an element fingerprint under a domain plus identifier.
- Relocate the same logical element after markup changes.
- Persist adaptive fingerprints with a schema-versioned `FileStore`.
- Build a `Response` from fetched or fixture content with HTTP metadata,
  selector behavior, body bytes, text, and JSON decoding.
- Fetch local/static pages through a basic `Fetcher` with GET, POST, PUT, and
  DELETE over `net/http`.
- Reuse a `FetcherSession` with default headers, per-request header overrides,
  cookie persistence, and connection reuse.
- Apply static fetcher redirect controls, request timeouts, retry attempts, and
  deterministic fetcher errors.
- Use an engine-neutral `BrowserFetcher` contract for dynamic page fetches,
  waits, page actions, and resource blocking.
- Run a fixture-backed spider core with request fingerprints, priority
  scheduling, duplicate skipping, callback output, follow requests, and named
  session routing.
- Use `goscrapling extract get/post/put/delete` for local/static page
  extraction with headers, timeout parsing, query params, request bodies,
  JSON bodies, CSS-selected text output, and full HTML output.
- Drive all behavior with test-first fixtures.

Missing major Scrapling subsystems:

- Advanced fetcher options such as form helpers, query params, response cookie
  access, redirect history, and browser-style header generation.
- Browser-backed fetching.
- Proxy rotation.
- Production crawler controls such as allowed domains, robots, cache,
  checkpoints, richer stats, and polite concurrency.
- Advanced CLI output modes, shell, MCP, or Gormes/OpenClaw tool integration.

The current adaptive parser is phase 0. It is not enough for the project to be considered a real Scrapling-style port.

## Port Target

The parity target is documented in:

- [Scrapling Parity Matrix](docs/research/scrapling-parity-matrix.md)
- [True Port Design](docs/superpowers/specs/2026-05-13-goscrapling-true-port-design.md)
- [Scrapling Feature Map](docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md)
- [Upstream Coverage Ledger](docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md)
- [Progress Ledger](docs/content/building-goscrapling/architecture_plan/progress.json)
- [Agent Queue](docs/content/building-goscrapling/builder-loop/agent-queue.md)
- [Portfolio and Gormes Fit](docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md)

Next required milestone:

1. Work the generated P0 rows for adaptive selector modes, selector extraction
   helpers, XPath/translator parity, and static fetcher options.
2. Keep using the generated agent queue; the upstream feature inventory is now
   split into builder-sized rows in `progress.json`.
3. Continue updating the parity matrix as each Scrapling subsystem moves from
   planned to partial or done.

Progress control:

```sh
go run ./cmd/progress validate
go run ./cmd/progress write
```

`validate` checks the canonical progress ledger. `write` regenerates the
builder-loop handoff, agent queue, next slices, blocked slices, and umbrella
cleanup pages from `progress.json`.

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
- [Portfolio and Gormes Fit](docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md)
- [Progress Schema](docs/content/building-goscrapling/builder-loop/progress-schema.md)
- [Agent Queue](docs/content/building-goscrapling/builder-loop/agent-queue.md)
- [Go Scraping OSS Survey](docs/research/go-scraping-oss-survey.md)
- [Adaptive Parser MVP Design](docs/superpowers/specs/2026-05-13-goscrapling-adaptive-parser-design.md)
- [True Port Design](docs/superpowers/specs/2026-05-13-goscrapling-true-port-design.md)
