# goscrapling Scrapling Taxonomy Design

## Goal

Refactor the Go source tree so the package layout resembles Scrapling's
feature taxonomy while preserving the current root `goscrapling` API as a
compatibility facade.

## Upstream Taxonomy

The observed upstream Scrapling checkout is organized around these runtime
areas:

- `scrapling/parser.py` for parser, selector, traversal, and extraction
  behavior.
- `scrapling/core/custom_types.py`, `storage.py`, and `translator.py` for
  reusable parser support.
- `scrapling/engines/static.py`, `engines/toolbelt/**`, and
  `engines/_browsers/**` for response conversion, static engine behavior,
  proxy/navigation helpers, and browser machinery.
- `scrapling/fetchers/**` for user-facing fetchers and sessions.
- `scrapling/spiders/**` for crawler requests, sessions, scheduler, results,
  and templates.
- `scrapling/cli.py` for command surfaces.

## Go Package Layout

Use Go package names that map to those upstream concepts without copying Python
private file names:

```text
core/customtypes/     TextHandler, TextHandlers, AttributesHandler
core/storage/         Key, Store, MemoryStore, FileStore, SQLiteStore, Fingerprint scoring
parser/               Document, Element, Selection, CSS/XPath, traversal, adaptive relocation
engines/toolbelt/     Response and response metadata/conversion helpers
engines/browser/      BrowserFetcher, BrowserEngine, chromedp adapter
fetchers/             Static Fetcher, FetcherSession, request options, proxy/error behavior
spiders/              Crawler, Request, Response, Result, Scheduler, SessionManager
cmd/goscrapling/      CLI entrypoint
internal/cli/         Go CLI implementation detail
internal/progress/    Progress-ledger tooling
```

The root package remains as `github.com/TrebuchetDynamics/goscrapling` and
re-exports the stable public types and constructors. Existing callers can keep
using `goscrapling.Parse`, `goscrapling.Fetcher`, `goscrapling.Response`, and
store constructors while newer code can import taxonomy packages directly.

## Compatibility Rules

- Do not copy upstream Python internals or private names directly.
- Do not move behavior without moving its tests or adding an equivalent package
  test.
- Keep `spiders` plural to match upstream `scrapling/spiders/**`.
- Keep the root facade until a future progress row explicitly approves a
  breaking public API migration.
- Keep browser, proxy, crawler, and integration behavior subject to the existing
  safety and fixture-test boundaries.

## Validation

The refactor is complete when the package graph builds and the repo validation
commands pass:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```
