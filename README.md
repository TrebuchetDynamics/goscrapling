# goscrapling

Go-native Scrapling-style web extraction for agent runtimes and
single-binary Go deployments.

Built for developers who want parser, selector, fetcher, browser, and spider
primitives in one Go module, without a Python sidecar.

`goscrapling` is an independent Go feature port inspired by
[D4Vinci/Scrapling](https://github.com/D4Vinci/Scrapling). It is useful today
as a tested parser, selector, static fetcher, response, browser-fetching seam,
and spider foundation. It is not a complete Scrapling port yet.

This project is not affiliated with D4Vinci/Scrapling. Scrapling is used as
public reference material for feature mapping and compatibility planning.

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

## Why goscrapling?

- No Python sidecar for Go agents, CLIs, or services.
- Fits single-binary Go deployments.
- Provides tested parser, selector, response, fetcher, and spider primitives.
- Tracks progress against Scrapling-visible behavior instead of relying on
  vague parity claims.
- Is designed for Gormes/OpenClaw-style agent runtimes.

## What Works Today

| Area | Current surface | Status |
|---|---|---|
| Parser | Parse HTML into `Document`, `Element`, and `Selection` values | Partial: core static parsing only |
| Selectors | CSS, XPath, `::text`, `::attr(name)`, extraction helpers, regex, JSON helpers | Partial: selector generation remains planned |
| Adaptive relocation | Save fingerprints and relocate logical elements after markup changes | Phase 0 foundation |
| Storage | In-memory, file-backed, and SQLite-backed adaptive stores | Partial: compatibility migration coverage is narrow |
| Response | HTTP metadata, headers, cookies, redirect history, body, text, JSON, CSS, XPath | Partial: later browser-produced XHR capture remains planned |
| Static fetcher | GET, POST, PUT, DELETE, sessions, headers, cookies, params, data, JSON, auth, verify, redirects, timeout, retry, explicit proxy options | Partial: impersonation, HTTP/3, and concurrent APIs remain planned |
| Browser fetcher | Engine-neutral contract plus chromedp-backed JavaScript rendering | Partial: basic rendering only; no stealth, proxy rotation, or browser-pool lifecycle yet |
| Spider | Fixture-backed scheduler, sessions, duplicate skipping, callbacks, follow requests, allowed domains, basic concurrency controls | Partial: robots, cache, checkpoints, retries, and templates remain planned |
| CLI | `goscrapling extract get/post/put/delete` for static extraction | Partial: advanced output modes, browser modes, and shell remain planned |

Major planned surfaces still include deeper browser behavior, proxy rotation,
production crawler controls, advanced CLI output, shell workflows, MCP, and
Gormes/OpenClaw tool integration.

## Quick Start

Requires Go 1.26 or newer. This follows the module's current `go` directive in
`go.mod`.

Use the module from a Go project:

```sh
go get github.com/TrebuchetDynamics/goscrapling
```

Or work from this checkout:

```sh
go test ./... -count=1
go run ./cmd/goscrapling help
```

### First Fetch

```go
package main

import (
	"fmt"

	"github.com/TrebuchetDynamics/goscrapling"
)

func main() {
	fetcher := goscrapling.Fetcher{}
	response, err := fetcher.Get("https://example.com", goscrapling.RequestOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(response.CSS("title::text").Get())
}
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

## Contributing

Feature work is tracked through a progress ledger and generated builder queue.
Start with [CONTRIBUTING.md](CONTRIBUTING.md) before adding behavior.

For local verification, run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
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
