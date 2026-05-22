# Upstream Coverage Ledger

The feature map says what Scrapling behavior should become in Go. This ledger
answers how we know the map is complete enough to plan from.

goscrapling does not claim parity by README prose. A planner pass can claim an
upstream surface is mapped only when every feature-bearing source class below
has a feature-map anchor, a Go target, and either a `progress.json` row or an
explicit owned/excluded decision. External projects such as Lightpanda may be
used as design references, but they do not replace Scrapling as the parity
oracle and must be cited as non-vendored references in `progress.json`.

Use [Scrapling Behavior Atoms](scrapling-behavior-atoms.md) for behavior-level
coverage. Use [Project Boundaries](boundaries.md) for owned Go-native package
and compatibility decisions.

## Completeness Standard

1. Inventory the upstream source class.
2. Link it to a feature-map anchor.
3. Name the Go target or owned replacement.
4. Make incomplete behavior reachable from `progress.json`.
5. When upstream changes, update this ledger, the feature map, and
   `progress.json` in the same planning pass.

## Audit Commands

Use this from the repo root when `references/Scrapling` is present:

```sh
find references/Scrapling/scrapling -maxdepth 3 -type f \
  \( -name '*.py' -o -name 'py.typed' \) \
  | sed 's#^references/Scrapling/##' \
  | awk -F/ '{print $1 "/" $2}' \
  | sort | uniq -c

find references/Scrapling/docs -maxdepth 2 -type f \
  \( -name '*.md' -o -name '*.png' -o -name '*.svg' -o -name '*.ico' \) \
  | sed 's#^references/Scrapling/##' \
  | awk -F/ '{print $1 "/" $2}' \
  | sort | uniq -c
```

The Go test `TestUpstreamCoverageLedgerMentionsCoreScraplingSources` checks the
core source-class names below when the local upstream checkout exists.

## Scrapling Source Coverage

| Upstream source class | Feature-map anchor | Go target | Progress anchor | Coverage |
|---|---|---|---|---|
| `scrapling/parser.py` | Parser and selector objects | `parser.Document`, `parser.Element`, `parser.Selection`, and root facade aliases | Phase 0 | partial |
| `scrapling/core/storage.py` | Adaptive storage | `core/storage.Store`, `MemoryStore`, `FileStore`, `SQLiteStore`, and root facade aliases | Phase 0, Phase 2 | partial |
| `scrapling/core/mixins.py` | Parser and selector objects | selector generation helpers with short/full CSS and XPath output | Phase 0 | covered |
| `scrapling/core/custom_types.py`, `scrapling/core/_types.py` | Parser and selector objects | `core/customtypes` plus root facade aliases for `TextHandler`, `TextHandlers`, and `AttributesHandler` behavior | Phase 0 | partial |
| `scrapling/core/translator.py` | Parser and selector objects | `core/translator.CSSToXPath` with `parser` and root facade aliases | Phase 0 | partial |
| `scrapling/core/shell.py`, `scrapling/core/_shell_signatures.py`, `scrapling/core/utils/_shell.py` | CLI shell and extract commands | `cmd/goscrapling`, `internal/cli` scripted shell evaluator, page shortcuts, and future curl helpers | Phase 5 | partial |
| `scrapling/core/utils/_utils.py` | Parser and adaptive storage | adaptive fingerprint field mapping and relocation diagnostics | Phase 0 | partial |
| `scrapling/core/ai.py` | MCP and AI docs | `integrations/mcp` | Phase 5 | planned |
| `scrapling/engines/static.py`, `scrapling/engines/toolbelt/custom.py`, `scrapling/engines/toolbelt/convertor.py` | Response object, static fetcher | `engines/toolbelt.Response`, `fetchers.Fetcher`, `fetchers.FetcherSession`, and root facade aliases | Phase 1 | partial |
| `scrapling/engines/_browsers/**` | Browser fetching | `engines/browser.BrowserFetcher`, `ChromedpBrowserEngine`, future sessions and page pools, and root facade aliases | Phase 3 | partial |
| `scrapling/engines/toolbelt/proxy_rotation.py` | Proxy rotation | `ProxyRotator`, fetcher/spider proxy options | Phase 1, Phase 4 | partial |
| `scrapling/engines/toolbelt/fingerprints.py` | Static fetcher, Browser fetching | explicit identity and stealth option boundaries | Phase 1, Phase 3 | planned |
| `scrapling/engines/toolbelt/navigation.py`, `scrapling/engines/toolbelt/ad_domains.py`, `scrapling/engines/constants.py` | Browser fetching, Proxy rotation | browser context/resource controls and proxy error helpers | Phase 1, Phase 3 | partial |
| `scrapling/fetchers/requests.py` | Static fetcher | `fetchers.Fetcher`, `fetchers.FetcherSession`, and root facade aliases | Phase 1 | partial |
| `scrapling/fetchers/chrome.py` | Browser fetching | `engines/browser.BrowserFetcher`, `ChromedpBrowserEngine`, and root facade aliases | Phase 3 | partial |
| `scrapling/fetchers/stealth_chrome.py` | Browser fetching, stealth controls | `browser` package plus explicit stealth options | Phase 3 future split | planned |
| `scrapling/spiders/**` | Spider runtime | `spiders` package request/result/session/scheduler contracts, allowed-domain filtering, crawler concurrency controls, and development response cache | Phase 4 | partial |
| `scrapling/cli.py` | CLI shell and extract commands | `cmd/goscrapling`, `internal/cli` | Phase 5 | partial |
| `Dockerfile`, `pyproject.toml`, `server.json` | Install, Docker, packaging, examples, and benchmarks | `cmd/goscrapling`, docs, integration metadata, README status map, public example tests, benchmark fixtures, and generated parity scorecard | Phase 5 | partial |
| `scrapling/py.typed` | Packaging/type marker | none | Coverage ledger | excluded |

## Scrapling Docs Coverage

| Upstream docs class | Feature-map anchor | Go target | Progress anchor | Coverage |
|---|---|---|---|---|
| `docs/parsing/**` | Parser and selector objects, adaptive storage | `parser`, `core/storage`, and root facade aliases | Phase 0, Phase 2 | partial |
| `docs/api-reference/selector.md`, `docs/api-reference/custom-types.md` | Parser and selector objects | `parser` selector helpers, selector generation methods, `core/customtypes`, and root facade aliases | Phase 0 | partial |
| `docs/api-reference/response.md` | Response object | `engines/toolbelt.Response` and root facade alias | Phase 1 | partial |
| `docs/api-reference/fetchers.md`, `docs/fetching/**` | Fetchers and browser fetching | `fetchers`, `engines/browser`, and root facade aliases | Phase 1, Phase 3 | partial |
| `docs/api-reference/proxy-rotation.md` | Proxy rotation | `ProxyRotator`, fetcher/browser/spider proxy options | Phase 1, Phase 4 | planned |
| `docs/spiders/**` | Spider runtime | `spiders` package with fixture-backed core, allowed-domain filtering, concurrency/domain-delay controls, and development response cache | Phase 4 | partial |
| `docs/cli/**` | CLI shell and extract commands | `cmd/goscrapling`, `internal/cli` | Phase 5 | partial |
| `docs/ai/mcp-server.md`, `docs/api-reference/mcp-server.md` | MCP and AI integration | `integrations/mcp` | Phase 5 | planned |
| `docs/benchmarks.md`, `docs/tutorials/**`, `docs/overview.md` | Install, Docker, packaging, examples, and benchmarks | docs, examples, benchmarks, README status map, public example tests, and generated parity scorecard | Phase 5 | partial |
| translated README files, assets, stylesheets, ReadTheDocs config | docs and branding | docs only | Coverage ledger | excluded |

## External Design References

| Reference | Observed ref | Used for | Progress rows | Boundary |
|---|---|---|---|---|
| `lightpanda-io/browser` | commit `5905319e78541110a0b4065b07ec7ce53f93a660` | Browser markdown dump, semantic tree, robots/private-network safety controls, and MCP-style tool taxonomy | `Browser fetch markdown dump`; `Browser semantic tree extraction`; `Fetch safety controls: robots and private network blocking`; `Gormes browser extraction tools` | Non-vendored design reference only; do not copy AGPL source into goscrapling without an explicit licensing decision. |

## What Counts As Unmapped

An upstream feature is unmapped when any of these are true:

- the source class is absent from this ledger;
- the ledger row points to no feature-map section;
- the feature map names no Go target or owned/excluded decision;
- incomplete behavior has no `progress.json` row;
- a row exists but has no source refs, write scope, tests or explicit
  no-test reason, acceptance, and done signal.

When this happens, run a `goscrapling-scrapling-parity` pass followed by a
`goscrapling-planner` pass. Do not hand the behavior to a builder until the
row is builder-ready.
