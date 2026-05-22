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
| Selection collections | `Selectors` class | partial | `Selection.Len`, `First`, `Text`, and `HTML` exist. Needs iteration, attr extraction, filtering, chaining, and get/getall-style helpers. |
| Element traversal | ancestors, parent, siblings, children | planned | Required to approach Scrapling parser ergonomics. |
| Selector generation | `core/mixins.py` | partial | `Element.GenerateCSSSelector`, `GenerateFullCSSSelector`, `GenerateXPathSelector`, and `GenerateFullXPathSelector` produce deterministic selectors that re-select fixture elements; deeper edge-case parity remains future work. |
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
| Pluggable storage | `StorageSystemMixin` | partial | `Store` interface and JSON `FileStore` exist with schema-version checks. SQLite parity remains planned. |
| SQLite default storage | `SQLiteStorageSystem` | planned | Required for useful production parity. |

## Fetcher And Response Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Response object | `engines/toolbelt/custom.py`, docs `fetching/choosing.md` | partial | Response now behaves like a parsed document with URL/status/header/request metadata, body bytes, text, and JSON decoding. Cookies, history, meta, and captured XHR remain future work. |
| Static HTTP fetcher | `fetchers/requests.py`, `engines/static.py` | partial | Basic GET, POST, PUT, DELETE, session defaults, cookies, connection reuse, redirects, timeouts, retry attempts, and error taxonomy are fixture-backed. Query params, form/JSON request helpers, browser-style headers, proxies, and response history remain planned. |
| Async HTTP fetcher | `AsyncFetcher` | planned | Go equivalent should use context and goroutines rather than Python-style awaitable API. |
| Fetcher sessions | `FetcherSession`, async sessions | partial | Synchronous Go-native session behavior has default headers, per-request overrides, cookies, and connection reuse. Async-style coordination remains future Go API work. |
| Request options merging | fetcher tests | partial | Session default headers and per-request overrides are covered. Broader request option merging remains planned. |
| Headers and browser-ish defaults | `toolbelt/fingerprints.py` | planned | Needs careful, honest implementation without overclaiming stealth. |
| Proxies and proxy rotation | `proxy_rotation.py`, docs `api-reference/proxy-rotation.md` | planned | Needed before spider production use. |
| Response history/cookies/body/status | fetcher docs/tests | partial | Body, status helpers, and session cookie persistence are covered; response cookie access and redirect history remain planned. |

## Browser Fetcher Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Dynamic browser fetcher | `fetchers/chrome.py`, docs `fetching/dynamic.md` | partial | Engine-neutral `BrowserFetcher` contract is fake-engine tested; real chromedp adapter is fixture-backed. Session pools, deeper wait/action behavior, XHR capture, and stealth remain planned. |
| Stealth browser fetcher | `fetchers/stealth_chrome.py`, docs `fetching/stealthy.md` | planned | Hard parity area; avoid unsupported anti-bot claims until proven. |
| Browser sessions | `engines/_browsers/` | planned | Required for reuse and spider sessions. |
| Wait selector/network idle | browser docs/tests | partial | Contract fields are covered through a fake engine; real browser fixture remains planned. |
| Captured XHR | response docs | planned | Important for modern web extraction. |
| Proxy/browser context settings | browser fetcher docs | planned | Later phase after basic browser fetch is stable. |
| Cloudflare challenge solving | stealth docs | deferred | High-risk claim; implement only after evidence and tests. |

## Spider/Crawler Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| Spider base type | `spiders/spider.py` | partial | Go-native `Crawler` and callback funcs exist with allowed-domain filtering and crawler concurrency controls; start URL helpers, lifecycle hooks, and streaming remain planned. |
| Request type | `spiders/request.py` | partial | URL, method/body, callback, priority, session ID, metadata, and dedupe flag are covered; retry/blocking metadata remains planned. |
| Response follow helpers | `engines/toolbelt/custom.py` | partial | Relative URL resolution, meta merge, callback/session override, priority, and referer flow are fixture-backed. Broader response ergonomics remain future work. |
| Scheduler | `spiders/scheduler.py` | done | Priority queue and duplicate filtering by deterministic request fingerprint are covered. Checkpoint snapshot/restore belongs to the checkpoint row. |
| Request fingerprinting | `spiders/request.py` | partial | URL/method/body/session fingerprints are deterministic with header and fragment options; kwargs/retry-related dimensions remain future work. |
| Session manager | `spiders/session.py` | partial | Named session routing plus eager/lazy startup is covered with fake sessions; static/dynamic/stealth session adapters remain planned. |
| Engine concurrency | `spiders/engine.py` | done | Global concurrency, per-domain active request limits, fixed download delay, backpressure, and context cancellation are covered with fake sessions. |
| Allowed domains | `spiders/engine.py` | done | Callback-yielded requests are filtered by exact host/subdomain match and offsite drops increment crawl stats. |
| Robots.txt | `spiders/robotstxt.py` | planned | Required for production-friendly crawling. |
| Development response cache | `spiders/cache.py` | planned | Useful for test/debug cycles. |
| Checkpoint pause/resume | `spiders/checkpoint.py` | planned | Required parity for long crawls. |
| Crawl result and stats | `spiders/result.py` | partial | Items, errors, skipped duplicates, offsite drops, request counts, configured concurrency controls, download delay, and per-session counts are covered; richer timing/status/cache/export stats remain planned. |
| Generic templates | `spiders/templates/` | planned | Later convenience layer. |

## CLI, Shell, AI, And Tooling Parity

| Scrapling Area | Upstream Reference | goscrapling Status | Notes |
| --- | --- | --- | --- |
| CLI fetch/extract | `cli.py`, docs `cli/` | partial | `goscrapling extract get/post/put/delete` is fixture-backed for local/static pages, headers, timeout, CSS selection, txt/html output, bodies, JSON, query params, and parse errors. Markdown, AI-targeted cleanup, and browser modes remain planned. |
| Interactive shell | `core/shell.py` | planned | Scripted shell command fixtures should come before a full REPL dependency. |
| MCP server | `core/ai.py`, docs `ai/mcp-server.md` | planned | Important if `goscrapling` becomes a Gormes/OpenClaw web-search tool. |
| Agent skill metadata | `agent-skill/` | deferred | Only after CLI/MCP shape is real. |
| Benchmarks | `benchmarks.py`, docs `benchmarks.md` | partial | Hermetic parser, static fetcher/response, spider scheduler, and CLI benchmark fixtures exist under `benchmarks/`, and `cmd/progress scorecard` generates `docs/research/parity-scorecard.md`; broader upstream benchmark methodology and real timing comparisons remain future work. |

## Implementation Order

1. Parser parity foundation: selection helpers, text/attr extraction, traversal, retrieve, adaptive selector options.
2. Response plus static fetcher: `Fetcher`, session, HTTP metadata, request option merging, local HTTP tests.
3. Durable adaptive storage: SQLite adapter and storage tests.
4. Browser fetcher depth: add sessions, deeper wait/action behavior, XHR capture, and context/resource controls on top of the chromedp adapter.
5. Spider core: request, scheduler, fingerprints, engine, stats, callbacks.
6. Robots/cache/checkpoint/session manager.
7. CLI fetch/extract.
8. MCP/Gormes integration.
9. Stealth browser features and proxy rotation once testable.

## Current Reality

`goscrapling` is far from parity. It currently covers only a small part of parser/adaptive behavior. That is acceptable only if the project continues toward this matrix. If the matrix is not the target, the project should not claim to be a real Scrapling-style port.
