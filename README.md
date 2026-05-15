# goscrapling

Go-native Scrapling-style web extraction for agent runtimes and
single-binary Go deployments.

`goscrapling` is an independent Go feature port inspired by
[D4Vinci/Scrapling](https://github.com/D4Vinci/Scrapling). It is useful today
as a tested parser, selector, static fetcher, response, browser-fetching seam,
and spider foundation. It is not a complete Scrapling port yet.

This project is not affiliated with D4Vinci/Scrapling. Scrapling is used as
public reference material and as the parity oracle for planning.

## Who This Is For

- Go developers who want Scrapling-style HTML extraction without a Python
  sidecar.
- Agent-runtime builders who need deterministic parser, fetcher, and future
  browser/crawler primitives inside a single binary.
- Maintainers and reviewers who want to see a large port broken into visible,
  tested, builder-sized slices.

If you need a mature, production-ready Scrapling replacement today, use
Scrapling. If you want a Go-native extraction core that is moving toward
Scrapling-visible behavior with tests and progress ledgers, this repository is
the workbench.

## What Works Today

| Area | Current surface | Status |
|---|---|---|
| Parser | Parse HTML into `Document`, `Element`, and `Selection` values | Partial |
| Selectors | CSS, XPath, `::text`, `::attr(name)`, extraction helpers, regex, JSON helpers | Partial |
| Adaptive relocation | Save fingerprints and relocate logical elements after markup changes | Phase 0 foundation |
| Storage | In-memory, file-backed, and SQLite-backed adaptive stores | Partial |
| Response | HTTP metadata, headers, cookies, redirect history, body, text, JSON, CSS, XPath | Partial |
| Static fetcher | GET, POST, PUT, DELETE, sessions, headers, cookies, params, data, JSON, auth, verify, redirects, timeout, retry, explicit proxy options | Partial |
| Browser fetcher | Engine-neutral contract plus chromedp-backed JavaScript rendering | Partial |
| Spider | Fixture-backed scheduler, sessions, duplicate skipping, callbacks, follow requests, allowed domains, basic concurrency controls | Partial |
| CLI | `goscrapling extract get/post/put/delete` for static extraction | Partial |

Major planned surfaces still include deeper browser behavior, proxy rotation,
production crawler controls, advanced CLI output, shell workflows, MCP, and
Gormes/OpenClaw tool integration.

## Quick Start

Requires Go 1.26 or newer.

Use the module from a Go project:

```sh
go get github.com/TrebuchetDynamics/goscrapling
```

Or work from this checkout:

```sh
go test ./... -count=1
go run ./cmd/goscrapling help
```

### Parse And Select

```go
package main

import (
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
)

func main() {
	doc, err := goscrapling.Parse(strings.NewReader(`
		<main>
			<article class="product" data-sku="A-42">Trail pack</article>
		</main>
	`), goscrapling.ParseOptions{URL: "https://example.com/products"})
	if err != nil {
		panic(err)
	}

	fmt.Println(doc.CSS(".product::text").Get())
	fmt.Println(doc.CSS(".product::attr(data-sku)").Get())
}
```

### Relocate A Changed Element

```go
ctx := context.Background()
store := goscrapling.NewMemoryStore()

before, _ := goscrapling.Parse(strings.NewReader(
	`<article class="product" id="p1">Product 1</article>`,
), goscrapling.ParseOptions{
	URL:   "https://example.com/products",
	Store: store,
})
element, _ := before.CSS("#p1").First()
_ = before.Save(ctx, element, "featured-product")

after, _ := goscrapling.Parse(strings.NewReader(
	`<article class="product" data-id="p1"><span>Product 1</span></article>`,
), goscrapling.ParseOptions{
	URL:   "https://example.com/products",
	Store: store,
})
match, ok, _ := after.Relocate(ctx, "featured-product")
fmt.Println(ok, match.Element.Text())
```

### Fetch And Extract

```go
fetcher := goscrapling.Fetcher{}
response, err := fetcher.Get("https://example.com/products", goscrapling.RequestOptions{
	Headers: http.Header{"User-Agent": []string{"goscrapling-example"}},
})
if err != nil {
	panic(err)
}

fmt.Println(response.StatusCode())
fmt.Println(response.CSS("title::text").Get())
```

### Use The CLI

```sh
go run ./cmd/goscrapling extract get https://example.com page.txt \
  --css-selector "body"

go run ./cmd/goscrapling extract post https://example.com/search result.html \
  --json '{"q":"scraping"}' \
  -H "Accept: text/html"
```

The CLI writes text for `.txt`, `.md`, or extensionless outputs, and HTML for
`.html` or `.htm` outputs.

## Project Boundary

Recommended public positioning:

> goscrapling is a Go-native web extraction engine inspired by Scrapling, built
> for agent runtimes and single-binary deployments.

Avoid treating it as any of these until the progress ledger and tests prove the
claim:

- a complete Scrapling clone;
- a drop-in Scrapling replacement;
- a production stealth scraper;
- a Cloudflare bypass or anti-bot system.

Browser, proxy, crawler, and tool-integration work must stay explicit,
operator-visible, and fixture-tested before it is documented as available
behavior.

## How Development Is Tracked

This repo uses Scrapling as the parity oracle and keeps implementation work
tied to visible planning artifacts:

- [Scrapling Feature Map](docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md)
- [Upstream Coverage Ledger](docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md)
- [Progress Ledger](docs/content/building-goscrapling/architecture_plan/progress.json)
- [Agent Queue](docs/content/building-goscrapling/builder-loop/agent-queue.md)

Before adding behavior, update or pick a builder-sized row in `progress.json`
with source refs, write scope, tests, acceptance, and a done signal. Do not
create a parallel backlog.

Regenerate queue surfaces after progress-ledger edits:

```sh
go run ./cmd/progress write
```

## Validation

Run the full hermetic validation suite before claiming a slice is complete:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Run the deterministic full local CLI smoke suite:

```sh
go test ./cmd/goscrapling -run TestGoscraplingFullLocalEndToEnd -count=1
```

Run the optional live practice-site suite only when live network access and
robots.txt preflight are acceptable:

```sh
GOSCRAPLING_LIVE_E2E=1 go test ./cmd/goscrapling -run TestLivePracticeSitesEndToEnd -count=1 -timeout 10m
```

## Reference Material

The upstream Scrapling repository is cloned locally for study at:

```text
references/Scrapling
```

The checkout is ignored by git. Public documentation records observed
architecture and decisions, not copied source.

## Documentation

- [Portfolio and Gormes Fit](docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md)
- [Scrapling Architecture Map](docs/research/scrapling-architecture-map.md)
- [Scrapling Parity Matrix](docs/research/scrapling-parity-matrix.md)
- [Progress Schema](docs/content/building-goscrapling/builder-loop/progress-schema.md)
- [Adaptive Parser MVP Design](docs/superpowers/specs/2026-05-13-goscrapling-adaptive-parser-design.md)
- [True Port Design](docs/superpowers/specs/2026-05-13-goscrapling-true-port-design.md)
