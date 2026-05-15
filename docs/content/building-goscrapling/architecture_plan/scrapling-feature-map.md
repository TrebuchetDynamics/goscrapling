# Scrapling Feature Map

This is the planner map for finishing goscrapling as a true Go-native
Scrapling-style feature port.

Hard rule: assume Scrapling-visible behavior is a parity target unless this map
or `progress.json` marks a narrow Go-native divergence. Simpler first Go code
is not an owned divergence by itself.

Use this page before splitting or building rows. Missing implementation work
must become or refine rows in `progress.json`; this page is not a side backlog.

Use [Scrapling Behavior Atoms](scrapling-behavior-atoms.md) for the
behavior-level parity checklist. Use [Project Boundaries](boundaries.md) when a
row needs a package-boundary or donor-compatibility decision.
Use [Portfolio and Gormes Fit](../strategy/portfolio-and-gormes-fit.md) for the
business boundary, public positioning, and downstream integration rationale.

## Reading Rules

- `covered` means goscrapling has repository evidence and tests.
- `partial` means the area has working code but not full Scrapling behavior.
- `planned` means `progress.json` has a row with source refs, write scope,
  tests, and acceptance.
- `vague` means the feature is represented only by umbrella prose and must be
  split before builder work.
- `owned` means goscrapling intentionally diverges with a Go-native contract.
- `excluded` means the upstream surface is repo hygiene or docs that do not
  change runtime behavior.

## Go Architecture Target

| Runtime concern | Go target | Proof gate |
|---|---|---|
| Static parsing and selectors | root package now, future `parser` package if split is needed | `go test ./... -run TestParseSelectsElementsWithCSS -count=1` |
| Adaptive fingerprints and relocation | root package now, future `adaptive` package if storage grows | `go test ./... -run 'TestSave|TestRelocate' -count=1` |
| Response and static fetching | root package or `fetcher` package | local `httptest` fixtures |
| Durable adaptive storage | root package plus store adapters | temp-directory persistence fixtures |
| Browser fetching | root package `BrowserFetcher`, future `browser` package if engine binding grows | fake browser engine contract before real engine |
| Spider runtime | `spider` package | fake fetcher scheduler fixtures |
| CLI | `cmd/goscrapling`, `internal/cli` | command-output fixtures |
| Gormes integration | `integrations/gormes` | fake tool-call fixtures |

## Product Fit

goscrapling is strategically useful when it keeps moving toward a tested
Scrapling-style extraction engine for Go agent runtimes. The portfolio story is
the porting method: upstream feature inventory, parity ledgers, generated
builder rows, and hermetic tests. The Gormes story is a future single-binary
web extraction substrate that can serve agent tools without a Python sidecar.

This map should not be used to imply completed parity. Public language should
say "Go-native web extraction engine inspired by Scrapling" until the relevant
rows are done.

## Scrapling Feature Map

| Scrapling feature | Upstream refs | goscrapling target | Implementation plan | Progress anchor | Status |
|---|---|---|---|---|---|
| Parser and selector objects | `scrapling/parser.py`, `scrapling/core/translator.py`, `scrapling/core/custom_types.py`, `docs/parsing/**`, `docs/api-reference/selector.md` | `Document`, `Response`, `Element`, `Selection`, future custom handler types | Basic CSS selection, XPath selection, scoped CSS-to-XPath translation, CSS pseudo-elements, extraction helpers, Go-native custom handler types, traversal, filtering, text/regex search, and similar-element lookup are covered. Planned rows still cover selector generation. | Phase 0 | partial |
| Adaptive storage | `scrapling/core/storage.py`, `docs/development/adaptive_storage_system.md`, `docs/parsing/adaptive.md` | `AdaptiveStore`, `MemoryStore`, `FileStore`, future `SQLiteStore` | Domain isolation, identifier lookup, fingerprint shape, deterministic relocation, adaptive selector modes, and relocation diagnostics are covered. Planned rows still cover SQLite parity and deeper durable storage behavior. | Phase 0, Phase 2 | partial |
| Response object | `docs/api-reference/response.md`, `scrapling/engines/toolbelt/custom.py`, `scrapling/engines/toolbelt/convertor.py` | `Response` | Metadata/body helpers, cookies, redirect history, meta, encoding, request/response header detail, and captured XHR attachment points are covered. Planned rows still cover later browser-produced XHR capture behavior. | Phase 1 | partial |
| Static fetcher | `docs/fetching/static.md`, `docs/api-reference/fetchers.md`, `scrapling/fetchers/requests.py`, `scrapling/engines/static.py` | `Fetcher`, `FetcherSession`, future concurrent fetcher | GET/POST/PUT/DELETE, sessions, redirects, timeout, retry, error taxonomy, and Scrapling-style params/data/JSON/auth/verify/cookie request options are covered. Planned rows cover proxy support, impersonation boundaries, HTTP/3 if supportable, and Go-native concurrency. | Phase 1 | partial |
| Proxy rotation | `docs/api-reference/proxy-rotation.md`, `scrapling/engines/toolbelt/proxy_rotation.py` | `ProxyRotator` plus fetcher/spider proxy options | Add explicit static proxy support first, then cyclic/custom rotation and spider integration. Avoid making proxy support implicit. | Phase 1, Phase 4 | planned |
| Browser fetching | `docs/fetching/dynamic.md`, `docs/fetching/stealthy.md`, `scrapling/fetchers/chrome.py`, `scrapling/fetchers/stealth_chrome.py`, `scrapling/engines/_browsers/**`, `scrapling/engines/toolbelt/ad_domains.py` | root package `BrowserFetcher`, future browser/session adapters | Engine-neutral contract is covered. Planned rows cover a real browser adapter, session/page pool, wait/actions/downloads/screenshots/XHR, context options, resource/ad blocking, stealth controls, and Cloudflare-claim boundaries. | Phase 3 | partial |
| Spider runtime | `docs/spiders/**`, `scrapling/spiders/**` | `spider` package | Request/result/session/scheduler contracts are fixture-backed. Planned rows cover engine concurrency, domain filters, robots, cache, checkpointing, blocked retries, lifecycle/streaming/stats/export, link extraction/templates, and session adapters. | Phase 4 | partial |
| CLI shell and extract commands | `scrapling/cli.py`, `docs/cli/**` | `cmd/goscrapling`, `internal/cli` | Static `extract get/post/put/delete` is fixture-backed. Keep advanced output modes, browser modes, and shell as separate rows. | Phase 5 | partial |
| MCP and AI docs | `docs/ai/mcp-server.md`, `docs/api-reference/mcp-server.md`, `scrapling/core/ai.py` | `integrations/mcp`, future integration packages | Treat as integration surfaces, not core parser behavior. Start with deterministic tool schemas, fake browser seams, and no live LLM dependency. | Phase 5 | planned |
| Install, Docker, packaging, examples, and benchmarks | `Dockerfile`, `pyproject.toml`, `docs/benchmarks.md`, `docs/tutorials/**`, `docs/overview.md` | `cmd/goscrapling`, `docs/`, `benchmarks/` | Track install/Docker utility behavior, dependency boundaries, examples, API docs, migration guidance, and parity scorecards with tests where executable. | Phase 5 | planned |
| Translated README files, assets, stylesheets, ReadTheDocs config | `docs/README_*.md`, `docs/assets/**`, `.github/**`, release metadata | docs only | Use as reference material and branding inputs, not runtime parity requirements. | Coverage ledger | excluded |

## Implementation Order

1. Close parser selector gaps: traversal, filtering, text/regex/similar search,
   and selector generation.
2. Close remaining static fetcher gaps: proxy support, proxy rotation,
   impersonation boundaries, HTTP/3 if supportable, and concurrency.
3. Build real browser support in layers: adapter, sessions, waits/actions,
   context/resource controls, then stealth boundaries.
4. Continue spider production controls in small rows: concurrency, domains,
   robots, cache, checkpoints, blocked retries, stats, templates, and session
   adapters.
5. Finish tool surfaces after core APIs are stable: advanced CLI, Gormes, MCP,
   install/Docker, examples, benchmarks, and parity scorecards.

## Current Reality

The repository is far from Scrapling parity. The phase 0 parser/adaptive work
is useful only as the foundation for fetchers, browser-backed fetching, and
spider behavior. Future planning should bias toward closing those gaps rather
than polishing the existing parser in isolation.

The project is still worth continuing for a company portfolio and Gormes if the
docs, tests, and progress ledger stay synchronized with what is actually
implemented. The next high-value integration proof is a static Gormes adapter
that returns structured evidence from local fixtures.
