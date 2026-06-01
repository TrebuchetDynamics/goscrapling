# Scrapling → goscrapling Mirror Matrix

Date: 2026-05-25

Purpose: mirror the D4Vinci/Scrapling feature surface into goscrapling targets,
current evidence, and remaining parity rows. This page is a human-readable
research matrix; `progress.json`, the feature map, the upstream coverage ledger,
and the generated upstream app map remain the authoritative planning surfaces.

## Parity Doctrine

`goscrapling` is a Go-native Scrapling-style feature port. It should preserve
Scrapling-visible behavior where practical, choose Go-native APIs where that is
clearer, and document intentional divergences. Upstream source is reference
material, not code to copy, and public docs must not imply affiliation with
D4Vinci/Scrapling.

## Source Snapshot

- Upstream repository: `https://github.com/D4Vinci/Scrapling.git`
- Local reference path: `references/Scrapling`
- Baseline commit used by the port ledger: `6380ef0f266a5fff898c18953d6b03ca320b2fd4`
- Baseline describe: `v0.4.8-1-g6380ef0`
- Upstream release observed: `v0.4.8`
- Remote check on 2026-05-25: `origin/main` at
  `ed008efcf77e13048192579d465b0ed2192d1925`; the delta from the baseline is a
  Docker-image command change in `docs/ai/mcp-server.md` plus a sponsor image,
  with no new feature-bearing Python source. The Docker command delta is covered
  by the planned `Install command, Docker image, and dependency packaging docs`
  row.

## Status Legend

Use the architecture-plan vocabulary:

- `covered`: repository evidence and tests exist for the named behavior.
- `partial`: working Go behavior exists, but the full Scrapling surface is not
  yet proven.
- `planned`: `progress.json` has a builder-ready or umbrella row for the gap.
- `owned`: goscrapling intentionally diverges with a documented Go-native
  contract.
- `excluded`: the upstream surface is not a Go runtime/product parity target.

Current ledger posture:

- `progress.json`: 57 rows, 45 complete, 12 planned.
- `upstream-app-map.json`: 23 mirror entries, 18 partial, 2 planned, 3 excluded.
- Validation command for the mirror inventory: `go run ./cmd/progress map-validate`.

## Feature Mirror Matrix

| Scrapling surface | Upstream refs | goscrapling target | Current evidence | Remaining parity anchor | Status |
|---|---|---|---|---|---|
| Package exports and public API shape | `scrapling/__init__.py`, package `__init__.py` files | Root facade plus `parser`, `core/storage`, `core/translator`, `fetchers`, `engines/browser`, `engines/toolbelt`, `spiders`, `integrations/*` packages | `Scrapling-shaped Go package taxonomy` is complete and root aliases preserve stable examples. | Future package/export deltas route through `Future upstream release delta and unclassified surfaces`. | partial |
| Parser, selectors, custom types, and translator | `scrapling/parser.py`, `scrapling/core/custom_types.py`, `_types.py`, `mixins.py`, `translator.py`, docs `parsing/**`, selector/custom-type API docs | `parser.Document`, `Element`, `Selection`, `core/customtypes`, `core/translator`, root aliases | Static parse/CSS, pseudo-elements, extraction helpers, XPath/CSS-to-XPath, traversal/filter/search/similar, selector generation, and custom handlers are tested. | No known open parser builder row; future deltas route through the upstream-delta umbrella before implementation. | partial |
| Adaptive storage and relocation | `scrapling/core/storage.py`, `core/utils/_utils.py`, docs `parsing/adaptive.md`, `development/adaptive_storage_system.md` | `core/storage.Store`, `MemoryStore`, `FileStore`, `SQLiteStore`, parser adaptive selectors and diagnostics | Domain isolation, save/retrieve, relocation scoring, adaptive selector fallback/auto-save, diagnostics, file store, and SQLite store rows are complete. | Future storage compatibility deltas route through the upstream-delta umbrella. | partial |
| Response object | `engines/toolbelt/custom.py`, `convertor.py`, `docs/api-reference/response.md` | `engines/toolbelt.Response` plus root facade | Metadata/selectors, body/text/bytes/JSON, cookies, history, meta, and captured-XHR attachment points are covered. | Browser-produced capture paths are covered through browser rows; future response deltas route through the upstream-delta umbrella. | partial |
| Static fetcher and sessions | `fetchers/requests.py`, `engines/static.py`, docs `fetching/static.md`, `api-reference/fetchers.md` | `fetchers.Fetcher`, `FetcherSession`, `ConcurrentFetcher` | Methods, sessions, cookies, redirects, timeout/retry/errors, params/forms/JSON/auth/verify/cookies, safety controls, and bounded concurrent fetching are tested. | Future static fetcher work should cite a specific upstream delta instead of copying Python async shape. | partial |
| Proxy rotation and network helper constants | `engines/toolbelt/proxy_rotation.py`, `engines/constants.py`, docs `api-reference/proxy-rotation.md` | `fetchers.ProxyRotator`, static proxy options, and spider proxy adapters | Static proxy support, proxy error classification, cyclic/custom rotators, session integration, retry-on-proxy-error rotation, and spider static/browser proxy rotation metadata are covered. | Future proxy deltas route through the upstream-delta umbrella. | partial |
| Identity, impersonation, and safety boundaries | `engines/toolbelt/fingerprints.py`, `navigation.py`, docs `fetching/stealthy.md` | Explicit identity options, safety controls, and honest unsupported-feature errors | Browser-like headers are opt-in; TLS/browser impersonation and HTTP/3 return unsupported errors; robots/private-network blocking is covered for fetchers. | Stealth and Cloudflare claims stay separate under browser rows. | owned |
| Browser fetchers and sessions | `fetchers/chrome.py`, `engines/_browsers/**`, docs `fetching/dynamic.md` | `engines/browser.BrowserFetcher`, `ChromedpBrowserEngine`, `BrowserSession`, context/resource/wait/action/capture options | Engine contract, real chromedp JavaScript fixture, markdown dump, semantic tree, session/page pool, waits/actions/downloads/screenshots/XHR, context/resource options, stealth controls, and Cloudflare/Turnstile unsupported boundary are complete. | Future browser deltas route through upstream-delta rows. | partial |
| Stealth browser fetcher | `fetchers/stealth_chrome.py`, `engines/_browsers/_stealth.py`, docs `fetching/stealthy.md` | Explicit browser stealth controls with operator-visible limits | `BrowserStealthOptions` covers deterministic browser-like headers, Google referer opt-in, WebRTC/WebGL launch controls, canvas-noise script injection, and visible unsupported Cloudflare/Turnstile challenge-solving errors before engine work. | Future solver work requires local fixtures and explicit controls. | partial |
| Spider core and crawler engine | `spiders/request.py`, `result.py`, `scheduler.py`, `session.py`, `engine.py`, `spider.py`, docs `spiders/**` | `spiders` package request/result/session/scheduler/crawler APIs | Request/result/session/scheduler contracts, response follow helpers, allowed-domain filtering, concurrency, per-domain limits, download delay, cancellation, cache fixtures, blocked response retry hooks, checkpoint pause/resume, robots.txt manager/delay directives, lifecycle hooks, streaming, item hooks, expanded stats, ItemList export, LinkExtractor, CrawlSpider/SitemapSpider helpers, and static/browser/stealth/proxy session adapters are covered. | Future spider deltas route through the upstream-delta umbrella. | partial |
| Spider production controls | `spiders/robotstxt.py`, `cache.py`, `checkpoint.py` | Robots manager, development cache, checkpoint store | Development response cache, robots.txt manager/delay directives, blocked retry hooks, checkpoint pause/resume, lifecycle hooks, streaming, and stats/export controls are complete. | Link/template and session-adapter rows. | partial |
| Link extraction and spider templates | `spiders/links.py`, `spiders/templates/**`, docs `spiders/generic-templates.md` | `spiders.LinkExtractor` plus `spiders/templates` crawler/sitemap helpers | Link extraction filters, crawl rules, sitemap indexes, robots Sitemap directives, alternate links, and request generation are fixture-backed. | Future template deltas route through the upstream-delta umbrella. | partial |
| CLI extract commands and shell | `scrapling/cli.py`, `core/shell.py`, `_shell_signatures.py`, `core/utils/_shell.py`, docs `cli/**` | `cmd/goscrapling`, `internal/cli` | Non-mutating install guidance, static extract GET/POST/PUT/DELETE, request bodies, Markdown output, AI-targeted cleanup, fake-backed dynamic/stealth browser command wiring, full local CLI E2E, cross-layer E2E, scripted shell get/post/put/delete shortcuts, and curl helper parsing/execution are covered. | `Shell scripted help and namespace listing`; future full REPL/deeper shell gaps require split rows. | partial |
| Gormes, MCP, and AI tools | `core/ai.py`, docs `ai/mcp-server.md`, `api-reference/mcp-server.md`, `server.json` | `integrations/gormes`, `integrations/mcp` | Static `web_extract`, browser extraction Gormes tools, and deterministic MCP-style get/bulk_get/fetch/bulk_fetch/stealthy_fetch/bulk_stealthy_fetch/screenshot/session tools are fixture-backed with fake seams and no live LLM dependency. | Future MCP transport/deployment deltas route through upstream-delta rows. | partial |
| Public docs, examples, benchmarks, packaging | `docs/overview.md`, `docs/tutorials/**`, `docs/benchmarks.md`, `Dockerfile`, `pyproject.toml` | README status map, examples, benchmarks, generated scorecards, install/Docker docs | README maps major upstream groups, examples compile, hermetic benchmark fixtures exist, `cmd/progress scorecard` writes `parity-scorecard.md`, upstream app map validates, and install-packaging docs describe the no-download browser/Docker boundary. | Future upstream release delta umbrella or specific published-packaging row. | partial |
| Translated READMEs, docs assets, stylesheets, sponsor images | `docs/README_*.md`, `docs/assets/**`, top-level images and branding docs | Docs/branding reference only | Coverage ledger marks translated docs/assets as excluded from runtime parity. | None. | excluded |

## Source, Test, And Docs Mirror Index

This table mirrors the generated upstream app map at a higher level. Use
`docs/content/building-goscrapling/architecture_plan/upstream-app-map.json` for
full upstream ref lists.

| Upstream mirror entry | goscrapling feature target | Progress evidence | Coverage |
|---|---|---|---|
| Package metadata and public exports | Root facade plus package-level adapters | Parser, response, fetcher, browser, spider, CLI, taxonomy rows | partial |
| Python typing marker | No Go target | Explicit exclusion | excluded |
| Parser/selectors/custom types/translator | Parser and selector objects | Parse, extraction, XPath, traversal/search, selector generation rows | partial |
| Adaptive storage and relocation helpers | Adaptive storage and relocation | Save, relocate, fallback/auto-save, diagnostics, file/SQLite store rows | partial |
| Static response and fetchers | Response and static fetcher | Response, static methods, sessions, request options, concurrent fetch rows | partial |
| Proxy rotation, impersonation, network constants | Proxy rotation and identity boundaries | Static proxy, rotator, identity boundary rows | partial |
| Browser and stealth fetchers | Browser fetching | Browser adapter/session/wait/action/context/stealth rows complete; Cloudflare row planned | partial |
| Spider core and crawler engine | Spider runtime | Core spider, concurrency, allowed-domain, robots, blocked retry, cache, checkpoint, lifecycle, streaming, stats, link/template, and session-adapter rows are complete | partial |
| Robots/cache/checkpoint controls | Spider runtime | Robots manager, cache, blocked retry, and checkpoint fixture coverage complete | partial |
| Link extraction and templates | Spider runtime | LinkExtractor plus CrawlSpider/SitemapSpider fixture coverage complete | partial |
| CLI extract and shell | CLI shell and extract commands | Install guidance, static extract, advanced output/browser command seams, shell get/post/put/delete page shortcuts, and curl helpers complete; shell help/namespace listing planned before deeper shell/REPL gaps | partial |
| MCP and AI tools | MCP/AI integration | Deterministic MCP tool schemas and fake static/browser/session fixtures complete; Gormes rows complete | partial |
| Upstream parser/adaptive tests | Parser/adaptive fixtures | Parser/adaptive rows cover local hermetic equivalents | partial |
| Upstream fetcher/browser tests | Fetcher/browser fixtures | Static/proxy/browser rows cover local hermetic equivalents | partial |
| Upstream spider tests | Spider fixtures | Core/concurrency/domain/cache/blocked-retry/checkpoint/robots/lifecycle-stats/link-template rows complete; remaining rows planned | partial |
| Upstream CLI/core/storage/AI tests | CLI, storage, MCP fixtures | CLI static/advanced extract, shell, storage, and MCP fixture rows complete | partial |
| Upstream parser/adaptive docs | Parser/adaptive docs and examples | Public docs/examples plus parser/adaptive rows | partial |
| Upstream response/fetching/browser docs | Response, fetcher, proxy, browser docs | Public docs/examples plus response/fetcher/browser rows | partial |
| Upstream spider docs | Spider docs and examples | Public docs/examples plus spider rows | partial |
| Upstream CLI/MCP/AI docs | CLI, MCP, AI docs | CLI install/extract/shell and MCP tool rows complete; deeper docs remain delta-driven | partial |
| Public docs, tutorials, donate page, benchmarks | README/docs/examples/benchmarks/packaging | Public docs/examples, install-packaging boundary, and scorecard complete; published packaging remains delta-driven | partial |
| Translated README files | No runtime target | Explicit exclusion | excluded |
| Docs assets and branding images | No runtime target | Explicit exclusion | excluded |

## Remaining Planned Feature Rows

The matrix is not a side backlog. These are the current `progress.json` anchors
for known remaining gaps:

- `Shell scripted help and namespace listing`
- `Future upstream release delta and unclassified surfaces` (umbrella only; split
  before builder work)

## Validation Commands

Use these after changing the matrix or canonical ledgers:

```sh
go run ./cmd/progress map-validate
go run ./cmd/progress map-write
go run ./cmd/progress validate
go test ./... -count=1
git diff --check
```
