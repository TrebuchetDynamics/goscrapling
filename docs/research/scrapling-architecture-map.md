# Scrapling Architecture Map

Date: 2026-05-13

Purpose: map the upstream D4Vinci/Scrapling architecture into Go-friendly components before implementation.

## Source Snapshot

- Upstream repository: `https://github.com/D4Vinci/Scrapling`
- Local reference path: `references/Scrapling`
- Local commit: `6380ef0f266a5fff898c18953d6b03ca320b2fd4`
- Local describe: `v0.4.8-1-g6380ef0`
- Upstream release observed: `v0.4.8`, published 2026-05-11

The clone is reference-only and ignored by git. `goscrapling` should learn from the public design and behavior, avoid copying implementation code, and track feature parity in `docs/research/scrapling-parity-matrix.md`.

## Top-Level Subsystems

Scrapling is organized around these major areas:

- Parser and selector model: `scrapling/parser.py`, `scrapling/core/mixins.py`, `scrapling/core/translator.py`
- Adaptive storage: `scrapling/core/storage.py`, `scrapling/core/utils/_utils.py`
- HTTP fetchers: `scrapling/fetchers/requests.py`, `scrapling/engines/static.py`
- Browser fetchers: `scrapling/fetchers/chrome.py`, `scrapling/fetchers/stealth_chrome.py`, `scrapling/engines/_browsers/`
- Spider framework: `scrapling/spiders/engine.py`, `scheduler.py`, `request.py`, `session.py`, `checkpoint.py`, `cache.py`, `robotstxt.py`
- CLI and agent integration: `scrapling/cli.py`, `scrapling/core/ai.py`
- Tests: `tests/parser`, `tests/core`, `tests/fetchers`, `tests/spiders`, `tests/cli`, `tests/ai`

## Parser And Selection

Scrapling exposes a `Selector` abstraction for parsed HTML and individual elements. Selection supports:

- CSS selection.
- XPath selection.
- Text and regex-oriented search.
- Attribute and traversal helpers.
- Selector generation helpers.
- `Response` objects that behave like selectors plus HTTP metadata.

Go mapping:

- `Document` represents a parsed page.
- `Element` wraps a DOM node with helper methods.
- `Selection` is a slice-like collection of `Element`.
- CSS is the first target because Go has mature libraries for CSS-like DOM traversal.
- XPath can be a later phase unless a mature dependency is selected during planning.

## Adaptive Scraping

Scrapling adaptive scraping has two phases:

1. Save the first matching element under a domain plus identifier.
2. Later, retrieve the saved element properties and score current-page elements by similarity.

Saved element properties include:

- Element tag name.
- Element text.
- Element attributes and values.
- Sibling tag names.
- Path tag names.
- Parent tag name.
- Parent attributes and values.
- Parent text.

Go mapping:

- `Fingerprint` is a serializable struct holding these stable properties.
- `Store` is an interface with `Save(ctx, key, fingerprint)` and `Load(ctx, key)`.
- `MemoryStore` is the first test backend.
- `SQLiteStore` is a later production backend.
- `Relocator` scores all candidate nodes and returns the best match with score metadata.

## Fetchers

Scrapling uses three fetcher families:

- Static HTTP fetcher for normal requests.
- Dynamic browser fetcher for JavaScript-loaded pages.
- Stealth browser fetcher for more difficult targets.

Each fetcher returns a `Response` that carries selector behavior plus HTTP metadata such as status, headers, cookies, body, request headers, redirects, and captured XHR when browser-backed.

Go mapping:

- `fetch/http` can start with `net/http` and keep the interface small.
- Advanced HTTP impersonation can be evaluated later with `enetx/surf` or a lower-level TLS/client dependency.
- Browser fetching should be behind an interface so `rod` or `chromedp` can be swapped after a focused spike.

## Spider Framework

Scrapling spiders combine parser and fetchers into a Scrapy-like async crawler. The flow is:

1. Spider emits starting requests.
2. Scheduler queues requests by priority and filters duplicates using fingerprints.
3. Engine enforces concurrency, per-domain limits, download delay, and robots rules.
4. Session manager routes each request to a named session.
5. Response goes to the callback.
6. Callback yields items or follow-up requests.
7. Optional checkpointing and cache support resume or replay crawls.

Go mapping:

- This is not part of the first MVP.
- Future implementation should use goroutines, context cancellation, channels, and explicit backpressure.
- Request fingerprints and scheduler behavior are good candidates for TDD before networking.

## CLI And MCP

Scrapling includes a CLI and an MCP server. These are valuable later because `goscrapling` may become a Gormes/OpenClaw web search tool.

Go mapping:

- Keep core library APIs independent of CLI and MCP.
- Add CLI only after parser and fetcher APIs stabilize.
- Add MCP or Gormes tool integration after fetch and extraction behavior is proven.

## Porting Doctrine

`goscrapling` should be a true Go-native feature port of Scrapling, not a small scraper that merely borrows the adaptive-selector idea.

Keep:

- Adaptive element relocation concept.
- Domain plus identifier storage model.
- Parser and response as the central user-facing objects.
- Clear fetcher layers.
- Test-first behavior fixtures.
- Major Scrapling subsystems as the long-term parity target.

Change for Go:

- Prefer explicit interfaces over class-level global configuration.
- Prefer `context.Context` in public operations that can block.
- Prefer small packages and typed structs over dynamic argument dictionaries.
- Build concurrency through Go primitives instead of copying Python async structure.

The parity matrix is the source of truth for what remains missing.
