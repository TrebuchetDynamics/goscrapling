# Scrapling Parity Matrix

Date: 2026-05-13

Purpose: define what "true Scrapling-style Go port" means for `goscrapling`, track current coverage, and prevent the project from drifting into a small unrelated scraper.

## Parity Doctrine

`goscrapling` is a long-term Go-native feature port of D4Vinci/Scrapling.

This does not mean line-by-line translation or Python API mimicry where that would be awkward in Go. It means `goscrapling` should cover Scrapling's major user-visible capabilities, preserve the same scraping mental model, and document any intentional differences.

Hard rules:

- Feature parity is the product target.
- Go-native APIs are acceptable only when the Scrapling behavior remains recognizable.
- Upstream source is reference material, not code to copy.
- Every parity item needs tests before implementation.
- Public docs must not imply affiliation with D4Vinci/Scrapling.

## Source Snapshot

- Upstream repository: `https://github.com/D4Vinci/Scrapling`
- Local reference path: `references/Scrapling`
- Local commit: `6380ef0f266a5fff898c18953d6b03ca320b2fd4`
- Local describe: `v0.4.8-1-g6380ef0`
- Upstream release observed: `v0.4.8`

## Status Legend

- `done`: implemented and tested in `goscrapling`
- `partial`: started but materially incomplete
- `planned`: required for parity, not started
- `deferred`: parity-relevant but intentionally later
- `excluded`: intentionally out of scope, with rationale

## Core Parser And Selector Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Main selector/document object | `scrapling/parser.py`, docs `parsing/main_classes.md` | partial | `Document`, `Element`, and `Selection` exist, but only a small subset of Scrapling selector behavior is implemented. |
| CSS selection | `Selector.css`, tests `tests/parser/` | partial | Basic CSS selection exists through goquery. Needs pseudo extraction behavior such as text and attr helpers. |
| XPath selection | `Selector.xpath` | planned | Needs dependency evaluation and test fixtures. |
| Text search | `find_by_text`, docs selection pages | planned | Needs exact/contains/regex behavior decisions. |
| Similar element search | `find_similar`, parser tests | planned | Related to adaptive scoring but user-facing API is missing. |
| Selection collections | `Selectors` class | partial | `Selection.Len` and `First` exist. Needs iteration, text extraction, attr extraction, filtering, chaining, and get/getall-style helpers. |
| Element traversal | ancestors, parent, siblings, children | planned | Required to approach Scrapling parser ergonomics. |
| Selector generation | `core/mixins.py` | planned | Needed for CSS/XPath generation parity. |
| Custom types | `core/custom_types.py`, docs `api-reference/custom-types.md` | planned | Go equivalent should be typed wrappers, not dynamic Python types. |

## Adaptive Scraping Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Element fingerprint extraction | `core/utils/_utils.py`, adaptive docs | partial | Current fingerprint captures tag, text, attrs, parent, siblings, path. Needs parity review against upstream fields and scoring. |
| Save by identifier | `Selector.save` | done | `Document.Save(ctx, element, identifier)` exists. |
| Retrieve by identifier | `Selector.retrieve` | partial | Store API exists, but no public `Document.Retrieve` helper yet. |
| Relocate | `Selector.relocate` | partial | `Document.Relocate` exists with deterministic scoring. Needs richer diagnostics and parity against upstream behavior. |
| Auto-save from CSS/XPath selectors | `auto_save=True` | planned | Needed for recognizably Scrapling-style adaptive usage. |
| Adaptive selector fallback | `adaptive=True` | planned | Needed so failed selectors can automatically relocate saved elements. |
| Domain override | `adaptive_domain` | planned | Current domain derivation exists; explicit override is missing. |
| Pluggable storage | `StorageSystemMixin` | partial | `Store` interface exists. Durable adapters missing. |
| SQLite default storage | `SQLiteStorageSystem` | planned | Required for useful production parity. |

## Fetcher And Response Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Response object | `engines/toolbelt/custom.py`, docs `fetching/choosing.md` | planned | Must behave like a parsed document plus HTTP metadata. |
| Static HTTP fetcher | `fetchers/requests.py`, `engines/static.py` | planned | Next major implementation slice. |
| Async HTTP fetcher | `AsyncFetcher` | planned | Go equivalent should use context and goroutines rather than Python-style awaitable API. |
| Fetcher sessions | `FetcherSession`, async sessions | planned | Required for connection reuse, default config, and spider integration. |
| Request options merging | fetcher tests | planned | Needs typed options and tests for defaults versus per-request overrides. |
| Headers and browser-ish defaults | `toolbelt/fingerprints.py` | planned | Needs careful, honest implementation without overclaiming stealth. |
| Proxies and proxy rotation | `proxy_rotation.py`, docs `api-reference/proxy-rotation.md` | planned | Needed before spider production use. |
| Response history/cookies/body/status | fetcher docs/tests | planned | Response metadata must be part of phase 1 of fetcher work. |

## Browser Fetcher Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Dynamic browser fetcher | `fetchers/chrome.py`, docs `fetching/dynamic.md` | planned | Needs `rod` versus `chromedp` decision. |
| Stealth browser fetcher | `fetchers/stealth_chrome.py`, docs `fetching/stealthy.md` | planned | Hard parity area; avoid unsupported anti-bot claims until proven. |
| Browser sessions | `engines/_browsers/` | planned | Required for reuse and spider sessions. |
| Wait selector/network idle | browser docs/tests | planned | Test with local HTTP fixtures first. |
| Captured XHR | response docs | planned | Important for modern web extraction. |
| Proxy/browser context settings | browser fetcher docs | planned | Later phase after basic browser fetch is stable. |
| Cloudflare challenge solving | stealth docs | deferred | High-risk claim; implement only after evidence and tests. |

## Spider/Crawler Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Spider base type | `spiders/spider.py` | planned | Go design likely uses interfaces and callback funcs instead of subclassing. |
| Request type | `spiders/request.py` | planned | Must include URL, method/body, callback, priority, session ID, metadata, retry count. |
| Response follow helpers | `engines/toolbelt/custom.py` | planned | Needed for spider ergonomics. |
| Scheduler | `spiders/scheduler.py` | planned | Priority queue and duplicate filtering by fingerprint. |
| Request fingerprinting | `spiders/request.py` | planned | Must be deterministic and test-first. |
| Session manager | `spiders/session.py` | planned | Routes requests to static/dynamic/stealth sessions. |
| Engine concurrency | `spiders/engine.py` | planned | Go implementation should use context, worker pools, and backpressure. |
| Allowed domains | engine tests | planned | Required for crawler safety. |
| Robots.txt | `spiders/robotstxt.py` | planned | Required for production-friendly crawling. |
| Development response cache | `spiders/cache.py` | planned | Useful for test/debug cycles. |
| Checkpoint pause/resume | `spiders/checkpoint.py` | planned | Required parity for long crawls. |
| Crawl result and stats | `spiders/result.py` | planned | Required for observability. |
| Generic templates | `spiders/templates/` | planned | Later convenience layer. |

## CLI, Shell, AI, And Tooling Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| CLI fetch/extract | `cli.py`, docs `cli/` | planned | Should wrap fetcher/parser APIs after they stabilize. |
| Interactive shell | `core/shell.py` | deferred | Useful but not required before fetcher/spider parity. |
| MCP server | `core/ai.py`, docs `ai/mcp-server.md` | planned | Important if `goscrapling` becomes a Gormes/OpenClaw web-search tool. |
| Agent skill metadata | `agent-skill/` | deferred | Only after CLI/MCP shape is real. |
| Benchmarks | `benchmarks.py`, docs `benchmarks.md` | planned | Needed after comparable fetcher/parser surfaces exist. |

## Implementation Order

1. Parser parity foundation: selection helpers, text/attr extraction, traversal, retrieve, adaptive selector options.
2. Response plus static fetcher: `Fetcher`, session, HTTP metadata, request option merging, local HTTP tests.
3. Durable adaptive storage: SQLite adapter and storage tests.
4. Browser fetcher spike: choose `rod` or `chromedp`, implement dynamic fetch with local fixture tests.
5. Spider core: request, scheduler, fingerprints, engine, stats, callbacks.
6. Robots/cache/checkpoint/session manager.
7. CLI fetch/extract.
8. MCP/Gormes integration.
9. Stealth browser features and proxy rotation once testable.

## Current Reality

`goscrapling` is far from parity. It currently covers only a small part of parser/adaptive behavior. That is acceptable only if the project continues toward this matrix. If the matrix is not the target, the project should not claim to be a real Scrapling-style port.

