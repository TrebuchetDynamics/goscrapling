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
| Browser fetching | `browser` package behind small interfaces | fake browser engine contract before real engine |
| Spider runtime | `spider` package | fake fetcher scheduler fixtures |
| CLI | `cmd/goscrapling`, `internal/cli` | command-output fixtures |
| Gormes integration | `integrations/gormes` | fake tool-call fixtures |

## Scrapling Feature Map

| Scrapling feature | Upstream refs | goscrapling target | Implementation plan | Progress anchor | Status |
|---|---|---|---|---|---|
| Parser and selector objects | `scrapling/parser.py`, `docs/parsing/main_classes.md`, `docs/api-reference/selector.md` | `Document`, `Element`, `Selection` | Keep the first Go API small while mapping Scrapling selector navigation, filtering, attribute helpers, text normalization, and adaptive save behavior into separate rows. | Phase 0 | partial |
| Adaptive storage | `scrapling/core/storage.py`, `docs/development/adaptive_storage_system.md`, `docs/parsing/adaptive.md` | `AdaptiveStore`, `MemoryStore`, future file store | Preserve domain isolation, identifier lookup, fingerprint shape, and deterministic relocation before adding durable store formats. | Phase 0, Phase 2 | partial |
| Response object | `docs/api-reference/response.md`, `scrapling/engines/static.py`, `scrapling/fetchers/requests.py` | `Response` | Build a response type that combines parsed document behavior with URL, status, headers, body, encoding, and request metadata. | Phase 1 | planned |
| Static fetcher | `docs/fetching/static.md`, `docs/api-reference/fetchers.md`, `scrapling/fetchers/requests.py` | `Fetcher`, `FetcherSession` | Start with `net/http`, local fixtures, request option merging, cookies, redirects, timeout, and visible error classes. | Phase 1 | planned |
| Proxy rotation | `docs/api-reference/proxy-rotation.md`, `scrapling/engines/toolbelt/proxy_rotation.py` | fetcher option or separate proxy package | Add only after fetcher error taxonomy and session behavior are stable. Avoid making proxy support implicit. | Future split under Phase 1 or Phase 4 | vague |
| Browser fetching | `docs/fetching/dynamic.md`, `docs/fetching/stealthy.md`, `scrapling/fetchers/chrome.py`, `scrapling/fetchers/stealth_chrome.py`, `scrapling/engines/_browsers/**` | `browser` package | Define the interface with fake engine tests before choosing chromedp, Playwright, Rod, or another engine. Split stealth behavior from normal browser fetching. | Phase 3 | planned |
| Spider runtime | `docs/spiders/**`, `scrapling/spiders/**` | `spider` package | Build request/result/session/scheduler first. Split robots, cache, checkpoints, stats, sitemap templates, and pagination into later rows. | Phase 4 | planned |
| CLI shell and extract commands | `scrapling/cli.py`, `docs/cli/**` | `cmd/goscrapling`, `internal/cli` | CLI comes after response/fetcher stability so output fixtures reflect real APIs. | Phase 5 | planned |
| MCP and AI docs | `docs/ai/mcp-server.md`, `docs/api-reference/mcp-server.md`, `scrapling/core/ai.py` | future integration packages | Treat as integration surfaces, not core parser behavior. Start with deterministic tool schemas and no live LLM dependency. | Phase 5 | planned |
| Examples, translated docs, assets, release config | `docs/README_*.md`, `docs/assets/**`, `.github/**`, release metadata | docs only | Use as reference material, not runtime parity requirements. | Coverage ledger | excluded |

## Implementation Order

1. Finish `Response`, `Fetcher`, and `FetcherSession`. Without these,
   goscrapling remains an adaptive parser toy rather than a scraping framework.
2. Add durable adaptive storage after the fetcher contract is stable.
3. Design the browser fetcher interface with fake fixtures before adding a real
   browser dependency.
4. Build the spider core on top of fake fetchers before robots/cache/checkpoint
   controls.
5. Add CLI and Gormes integration only after core APIs are stable enough to
   support fixture-backed command and tool output.

## Current Reality

The repository is far from Scrapling parity. The phase 0 parser/adaptive work
is useful only as the foundation for fetchers, browser-backed fetching, and
spider behavior. Future planning should bias toward closing those gaps rather
than polishing the existing parser in isolation.
