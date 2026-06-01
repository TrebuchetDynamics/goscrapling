# goscrapling True Port Design

Date: 2026-05-13

Status: scope correction approved by Juan.

## Decision

`goscrapling` will be a true long-term Go-native feature port of D4Vinci/Scrapling.

The project is no longer framed as a small adaptive scraping experiment. The adaptive parser MVP remains useful, but it is only phase 0. The product is worth continuing only if it aims for Scrapling-style parity across parser, adaptive storage, fetchers, browser-backed fetching, spiders, CLI, and agent/tool integrations.

## Meaning Of "True Port"

True port means feature and behavior parity, not line-by-line translation.

`goscrapling` should:

- Preserve Scrapling's user mental model.
- Cover Scrapling's major subsystems.
- Translate Python-only patterns into idiomatic Go.
- Maintain a parity matrix that shows what is done, partial, planned, deferred, or excluded.
- Use tests as the source of truth for each parity claim.

`goscrapling` should not:

- Copy upstream source code.
- Claim official affiliation.
- Overclaim anti-bot, stealth, browser, crawler, or MCP support before those features exist.
- Stop at adaptive selection and still present itself as a scraping framework.

## Product Architecture

The full project should converge on these packages:

- Root package `goscrapling`: parser, document, element, selection, adaptive API, common options.
- `fetch`: static HTTP fetcher, response object, sessions, request options, proxy support.
- `browser`: dynamic browser fetcher and later stealth browser fetcher.
- `storage`: durable adaptive stores, starting with SQLite.
- `crawl`: spider engine, request, scheduler, fingerprints, session manager, robots, cache, checkpoint, stats.
- `cmd/goscrapling`: CLI commands for fetch, extract, shell-like workflows, and later spider execution.
- `tool`: MCP/Gormes/OpenClaw integration after core behavior is stable.

The root package can keep the current MVP code until package boundaries become necessary. Do not split prematurely, but do not let root become a dumping ground for fetcher and crawler code.

## Porting Strategy

Use a parity-driven loop:

1. Choose one upstream subsystem.
2. Read upstream docs, public API examples, and tests for that subsystem.
3. Write Go parity notes and fixtures.
4. Write failing Go tests that capture the behavior.
5. Implement the smallest Go-native API that passes those tests.
6. Update the parity matrix from `planned` to `partial` or `done`.
7. Commit and push the slice.

Each slice should be independently useful and testable.

## API Strategy

Use Go-native APIs where Python APIs do not map cleanly.

Examples:

- Python class-level fetcher configuration should become explicit option structs and sessions.
- Python async APIs should become context-aware functions, goroutines, and channels where useful.
- Python spider subclassing should become interfaces and callback functions.
- Dynamic dictionaries should become typed request/response structs.

Keep names recognizable where it helps:

- `Fetcher`
- `Response`
- `DynamicFetcher`
- `StealthyFetcher`
- `Spider`
- `Request`
- `Scheduler`
- `Store`
- `Relocate`

## Required Milestones

### Milestone 0: Adaptive Parser Foundation

Status: partial and already started.

Required before moving on:

- Expand selection API beyond `Len` and `First`.
- Add text and attribute extraction helpers.
- Add public retrieve helper.
- Add adaptive selector convenience matching Scrapling's `auto_save` and `adaptive` concepts.
- Update parity matrix as parser items move from partial to done.

### Milestone 1: Response And Static Fetcher

This is the next major feature milestone because a scraper without fetching is not a real scraping framework.

Deliverables:

- `Response` type that behaves like `Document` plus HTTP metadata.
- Static `Fetcher` using `net/http` first.
- `FetcherSession` for connection reuse and defaults.
- Request option merging with typed options.
- Status, reason, headers, cookies, body, URL, and request headers.
- Local HTTP test server fixtures.

### Milestone 2: Durable Adaptive Storage

Deliverables:

- SQLite store.
- Domain override equivalent to Scrapling's `adaptive_domain`.
- Store tests covering overwrite behavior, isolation, and concurrency.

### Milestone 3: Browser Fetching

Deliverables:

- Compare `rod` and `chromedp`.
- Implement `DynamicFetcher` behind a small browser interface.
- Support wait selector and network idle where practical.
- Return a `Response`.
- Test against local JS-rendered fixtures.

### Milestone 4: Spider Core

Deliverables:

- `Request`.
- Request fingerprinting.
- Priority scheduler with duplicate filtering.
- Basic engine with concurrency.
- Callback-based item and request yielding.
- Result items and stats.

### Milestone 5: Production Crawling Features

Deliverables:

- Session manager.
- Allowed domains.
- Robots.txt.
- Response cache.
- Checkpoints and resume.
- Proxy rotation.

### Milestone 6: CLI And Tool Integration

Deliverables:

- CLI fetch/extract commands.
- MCP server or Gormes/OpenClaw tool adapter.
- Documentation and examples.

### Milestone 7: Stealth Features

Deliverables:

- Stealth fetcher only after browser basics are stable.
- Honest feature claims backed by tests and reproducible checks.
- Proxy/browser context support.

## Next Slice Recommendation

The next implementation slice should be Milestone 1: Response and static `Fetcher`.

Reason:

- It changes `goscrapling` from a parser utility into an actual scraping library.
- It is testable with local HTTP fixtures.
- It creates the response object that browser fetchers and spiders will reuse.
- It keeps the project aligned with the true-port requirement.

## Acceptance Criteria

The true-port effort is on track only if:

- `docs/research/scrapling-parity-matrix.md` exists and stays current.
- README states that parity is the long-term goal and that current status is partial.
- Every new feature maps to an upstream Scrapling subsystem.
- Tests pass with `go test ./...`.
- Public claims never outrun implemented behavior.

