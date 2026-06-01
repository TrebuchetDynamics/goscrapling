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
| Spider | Fixture-backed scheduler, sessions, duplicate skipping, callbacks, follow requests, allowed domains, basic concurrency controls, development response cache | Partial: robots, checkpoints, blocked retries, exports, and templates remain planned |
| CLI | `goscrapling install`, static `extract get/post/put/delete`, fake-backed `extract fetch`/`stealthy-fetch` browser seams, Markdown/HTML/text outputs, AI-targeted cleanup, plus scripted `shell -c` `get`, `uncurl`, `curl2fetcher`, `page`, `response`, and `pages` shortcuts | Partial: additional shell method shortcuts, full REPL, richer shell expressions, and published packaging remain planned |
| Tool integration | static Gormes `web_extract` adapter with selector evidence from local fixtures | Partial: MCP server and browser-backed tool calls remain planned |

## Parity Status Map

Upstream Scrapling docs group the product around parsing, adaptive scraping,
fetchers, browser fetchers, spiders, terminal tools, MCP/AI tooling, and
migration guidance. goscrapling maps those groups as follows; this is a status
map, not a complete parity claim.

| Upstream feature group | Go status |
|---|---|
| Parser and selectors | Partial: static HTML parsing, CSS, XPath, pseudo-elements, extraction helpers, traversal, filtering, text/regex search, and similar-element lookup are fixture-backed; selector generation remains planned. |
| Adaptive relocation | Partial: in-memory, file, and SQLite-backed fingerprints plus deterministic relocation are fixture-backed; migration/deeper compatibility coverage remains planned. |
| Static fetcher and response | Partial: GET/POST/PUT/DELETE, sessions, request options, response metadata/body helpers, redirects, cookies, retry, timeout, and explicit proxy options are fixture-backed; impersonation and HTTP/3 remain planned. |
| Browser fetching | Partial: engine-neutral browser fetcher and chromedp rendering exist; stealth, browser pools, screenshots/downloads, XHR capture, and resource controls remain planned. |
| Spider runtime | Partial: request/result/session/scheduler contracts, allowed-domain filtering, concurrency/domain-delay controls, and development response cache are fixture-backed; robots, checkpoints, retries, exports, and templates remain planned. |
| CLI | Partial: install guidance, static `extract` commands, Markdown/HTML/text outputs, AI-targeted cleanup, fake-backed browser-mode command wiring, scripted `shell -c` page shortcuts, and curl helper parsing/execution are fixture-backed; additional shell method shortcuts, full REPL, richer shell expressions, and published packaging remain planned. |
| MCP and Gormes integration | Partial: static Gormes `web_extract` returns URL, title, content, status, content type, final URL, selector evidence, and engine/mode metadata from local fixtures; MCP and browser-backed tools remain planned. |
| Migration guidance | Partial: README examples and status tables map the stable Go API to upstream Scrapling concepts; complete BeautifulSoup/Scrapy/AI migration guides remain planned. |
| Translated README files, assets, stylesheets, and ReadTheDocs config | Excluded: reference and branding support material only, not Go runtime parity work. |

Major planned surfaces still include future upstream deltas, published packaging,
additional shell shortcuts, full REPL shell workflows, and deeper migration/API
reference docs.

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
go run ./cmd/goscrapling install
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

go run ./cmd/goscrapling extract fetch https://example.com/app rendered.md \
  --network-idle --ai-targeted
```

The CLI writes text for `.txt` or extensionless outputs, Markdown for `.md`,
and HTML for `.html` or `.htm` outputs. `goscrapling install` is
non-mutating: it prints Go install, browser-runtime, and Docker packaging
guidance instead of downloading browsers.

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

- [Changelog](CHANGELOG.md)
- [v0.1.2 Release Notes](docs/releases/v0.1.2.md)
- [v0.1.1 Release Notes](docs/releases/v0.1.1.md)
- [v0.1.0 Release Notes](docs/releases/v0.1.0.md)
- [Portfolio and Gormes Fit](docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md)
- [Scrapling Architecture Map](docs/research/scrapling-architecture-map.md)
- [Scrapling Parity Matrix](docs/research/scrapling-parity-matrix.md)
- [Progress Schema](docs/content/building-goscrapling/builder-loop/schema/progress-schema.md)
- [Adaptive Parser MVP Design](docs/superpowers/specs/parser-foundation/2026-05-13-goscrapling-adaptive-parser-design.md)
- [True Port Design](docs/superpowers/specs/port-architecture/2026-05-13-goscrapling-true-port-design.md)
