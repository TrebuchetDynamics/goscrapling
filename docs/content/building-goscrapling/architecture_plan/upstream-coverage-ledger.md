# Upstream Coverage Ledger

The feature map says what Scrapling behavior should become in Go. This ledger
answers how we know the map is complete enough to plan from.

goscrapling does not claim parity by README prose. A planner pass can claim an
upstream surface is mapped only when every feature-bearing source class below
has a feature-map anchor, a Go target, and either a `progress.json` row or an
explicit owned/excluded decision.

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
| `scrapling/parser.py` | Parser and selector objects | `Document`, `Element`, `Selection` | Phase 0 | partial |
| `scrapling/core/storage.py` | Adaptive storage | `AdaptiveStore`, `MemoryStore`, future file store | Phase 0, Phase 2 | partial |
| `scrapling/core/mixins.py`, `scrapling/core/custom_types.py`, `scrapling/core/_types.py` | Parser and selector objects | root package types, future typed helpers | Phase 0 future splits | planned |
| `scrapling/core/translator.py`, `scrapling/core/shell.py`, `scrapling/core/_shell_signatures.py` | CLI shell and extract commands | `cmd/goscrapling`, `internal/cli` | Phase 5 | planned |
| `scrapling/core/ai.py` | MCP and AI docs | future integration packages | Phase 5 | planned |
| `scrapling/engines/static.py` | Response object, static fetcher | `Response`, `Fetcher`, `FetcherSession` | Phase 1 | planned |
| `scrapling/engines/_browsers/**` | Browser fetching | `browser` package | Phase 3 | planned |
| `scrapling/engines/toolbelt/proxy_rotation.py`, `scrapling/engines/toolbelt/fingerprints.py`, `scrapling/engines/toolbelt/navigation.py`, `scrapling/engines/toolbelt/ad_domains.py` | Proxy rotation and browser support | fetcher/browser options | Future split under Phase 1 or Phase 3 | vague |
| `scrapling/fetchers/requests.py` | Static fetcher | `Fetcher`, `FetcherSession` | Phase 1 | planned |
| `scrapling/fetchers/chrome.py` | Browser fetching | `browser` package | Phase 3 | planned |
| `scrapling/fetchers/stealth_chrome.py` | Browser fetching, stealth controls | `browser` package plus explicit stealth options | Phase 3 future split | planned |
| `scrapling/spiders/**` | Spider runtime | `spider` package | Phase 4 | planned |
| `scrapling/cli.py` | CLI shell and extract commands | `cmd/goscrapling`, `internal/cli` | Phase 5 | planned |
| `scrapling/py.typed` | Packaging/type marker | none | Coverage ledger | excluded |

## Scrapling Docs Coverage

| Upstream docs class | Feature-map anchor | Go target | Progress anchor | Coverage |
|---|---|---|---|---|
| `docs/parsing/**` | Parser and selector objects, adaptive storage | root package | Phase 0, Phase 2 | partial |
| `docs/api-reference/selector.md`, `docs/api-reference/custom-types.md` | Parser and selector objects | root package types | Phase 0 future splits | planned |
| `docs/api-reference/response.md` | Response object | `Response` | Phase 1 | planned |
| `docs/api-reference/fetchers.md`, `docs/fetching/**` | Fetchers and browser fetching | `Fetcher`, `FetcherSession`, `browser` | Phase 1, Phase 3 | planned |
| `docs/api-reference/proxy-rotation.md` | Proxy rotation | fetcher/browser options | Future split | vague |
| `docs/spiders/**` | Spider runtime | `spider` package | Phase 4 | planned |
| `docs/cli/**` | CLI shell and extract commands | `cmd/goscrapling`, `internal/cli` | Phase 5 | planned |
| `docs/ai/mcp-server.md`, `docs/api-reference/mcp-server.md` | MCP and AI integration | future integration packages | Phase 5 | planned |
| translated README files, assets, stylesheets, ReadTheDocs config | docs and branding | docs only | Coverage ledger | excluded |

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
