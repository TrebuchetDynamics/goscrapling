# Changelog

All notable goscrapling release changes are recorded here.

goscrapling follows Go module tags for releases. The project is still a partial
Go-native Scrapling-style feature port; release notes must not claim complete
Scrapling parity until the progress ledger and tests prove it.

## v0.1.2 - 2026-06-01

Patch release for the `v0.1` line, focused on shipping the latest hermetic
Scrapling-parity surfaces through a stable Go module tag.

### Added

- Scripted CLI shell support for static `get/post/put/delete` shortcuts,
  `help()`, `pages` history inspection, curl conversion helpers, and bounded
  `view(page)`/`view(response)` artifact writing without launching a browser in
  tests.
- Browser, Gormes, and MCP fixture-backed surfaces for rendered markdown,
  semantic trees, browser sessions, wait/action/capture controls, and
  deterministic tool handlers.
- Spider runtime fixtures for robots, cache, checkpointing, blocked retries,
  lifecycle/stat/export behavior, link extraction, crawl/sitemap templates, and
  session adapter seams.
- Refreshed progress ledgers, upstream app-map evidence, builder queues,
  benchmark fixtures, and parity scorecard generation after package/topology
  refactors.

### Known gaps

- goscrapling remains a partial Scrapling-style port.
- Full Python/IPython REPL compatibility, arbitrary shell expressions, live
  browser deployment assumptions, and any challenge-solving or anti-bot bypass
  behavior remain outside this release.

### Validation for release candidate

Release candidate validation should run from a clean checkout:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

## v0.1.1 - 2026-05-20

Maintenance release after `v0.1.0`, focused on internal architecture deepening for future Scrapling-parity work. No new complete Scrapling surface is claimed in this release.

### Changed

- Deepened spider crawl orchestration behind a private crawl runtime while keeping the public `spiders.Crawler.Run` interface stable.
- Centralized browser-rendered evidence extraction so markdown, semantic nodes, links, structured data, and interactive nodes share one parsed HTML evidence module.
- Deepened the static CLI extract command around an internal extraction plan while preserving existing `goscrapling extract get/post/put/delete` behavior.

### Validation for release candidate

Release candidate validation should run from a clean checkout:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Optional live E2E validation, only when live network access and robots.txt
preflight are acceptable:

```sh
GOSCRAPLING_LIVE_E2E=1 go test ./cmd/goscrapling -run TestLivePracticeSitesEndToEnd -count=1 -timeout 10m
```

## v0.1.0 - 2026-05-20

Initial v0.1 release candidate for the tested goscrapling foundation.

### Added

- Parser and selector foundation:
  - static HTML parsing into `Document`, `Element`, and `Selection` values;
  - CSS and XPath selection;
  - `::text` and `::attr(name)` extraction;
  - extraction helpers for text, HTML, regex, JSON, and attributes;
  - traversal, filtering, text search, regex search, and similar-element helpers.
- Adaptive relocation foundation:
  - in-memory adaptive store;
  - file-backed adaptive store;
  - SQLite-backed adaptive store;
  - deterministic relocation scoring and diagnostics;
  - CSS adaptive selector fallback and auto-save modes.
- Response and static fetcher foundation:
  - response metadata, body, text, bytes, JSON, selectors, cookies, redirect history, request metadata, and meta fields;
  - static `GET`, `POST`, `PUT`, and `DELETE` fetcher methods;
  - fetcher sessions, cookies, headers, params, form bodies, JSON bodies, basic auth, TLS verify control, redirects, timeouts, retries, and explicit static proxy options;
  - opt-in fetch safety controls for robots/private-network boundaries.
- Browser foundation:
  - engine-neutral browser fetcher interface;
  - chromedp-backed JavaScript fixture support;
  - browser-rendered markdown dump;
  - browser semantic tree extraction.
- Spider foundation:
  - request/result/session/scheduler contracts;
  - callback and follow-request flow;
  - duplicate filtering;
  - allowed-domain filtering;
  - concurrency and domain-delay controls.
- CLI foundation:
  - `goscrapling extract get/post/put/delete`;
  - text and selected HTML output;
  - deterministic full local binary E2E tests;
  - opt-in robots-gated live practice-site E2E suite.
- Integration foundation:
  - Gormes browser extraction tools for rendered markdown, links, semantic tree, structured data, and interactive elements;
  - cross-layer local E2E harness covering fetcher, response selectors, parser adaptive storage, spider flow, CLI binary extraction, and Gormes browser adapter.
- Port governance:
  - Scrapling feature map;
  - upstream coverage ledger;
  - progress ledger and generated builder queue;
  - repo-local development skills for parity, planning, TDD slices, and builder workflow.

### Known gaps

- goscrapling is not a complete Scrapling port.
- Selector generation remains planned.
- Static fetcher impersonation, proxy rotation strategies, HTTP/3, and concurrent API surfaces remain planned.
- Browser sessions, page pools, waits/actions, screenshots, downloads, XHR capture, resource controls, stealth controls, and Cloudflare challenge boundaries remain planned.
- Spider robots, response cache, checkpoints, blocked retries, lifecycle hooks, streaming, expanded stats, link extraction, templates, and session adapters remain planned.
- Advanced CLI modes, interactive shell, MCP server, install/Docker packaging docs, public API docs, examples, benchmarks, and parity scorecards remain planned.

### Validation for release candidate

Release candidate validation should run from a clean checkout:

```sh
go test ./... -count=1
go test ./cmd/goscrapling -count=1 -v
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Optional live E2E validation, only when live network access and robots.txt
preflight are acceptable:

```sh
GOSCRAPLING_LIVE_E2E=1 go test ./cmd/goscrapling -run TestLivePracticeSitesEndToEnd -count=1 -timeout 10m
```
