# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Static proxy support and proxy error classification

- Phase: `phase-1-response-fetcher / static-fetcher`
- Priority: `P1`
- Owner: `fetcher`
- Size: `medium`
- Contract status: `draft`
- Contract: Add explicit per-request and per-session proxy configuration with visible proxy error classification before rotation is introduced.
- Ready when: Fetcher error taxonomy is validated.
- Write scope: `fetcher.go`, `fetcher_proxy.go`, `fetcher_proxy_test.go`
- Test commands: `go test ./... -run TestStaticFetcherProxySupport -count=1`
- Acceptance: Local proxy fixtures prove proxy routing, proxy auth, bypass errors, and operator-visible error kinds.
- Done signal: Static proxy support tests pass.
- Source refs: `references/Scrapling/scrapling/engines/static.py`, `references/Scrapling/scrapling/engines/toolbelt/navigation.py`, `references/Scrapling/docs/fetching/static.md`

## 2. SQLite adaptive store parity

- Phase: `phase-2-storage / persistent-store`
- Priority: `P1`
- Owner: `storage`
- Size: `medium`
- Contract status: `draft`
- Contract: Add a SQLite-backed adaptive store matching Scrapling default durable storage behavior while preserving the Go Store interface.
- Ready when: FileStore compatibility row is complete and Store interface remains stable.
- Write scope: `store_sqlite.go`, `store_sqlite_test.go`, `testdata/adaptive_store/`
- Test commands: `go test ./... -run TestSQLiteStore -count=1`
- Acceptance: Temp-database fixtures prove save/load, domain isolation, close behavior, and schema compatibility errors.
- Done signal: SQLite store tests pass with temp databases.
- Source refs: `references/Scrapling/scrapling/core/storage.py`, `references/Scrapling/docs/development/adaptive_storage_system.md`

## 3. Real browser adapter with JavaScript fixture

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `draft`
- Contract: Bind BrowserFetcher to one real Go browser engine and prove dynamic content extraction against a local JavaScript fixture.
- Ready when: BrowserFetcher interface row is validated.
- Write scope: `browser.go`, `browser_adapter.go`, `browser_integration_test.go`, `testdata/browser/`
- Test commands: `go test ./... -run TestBrowserAdapter -count=1`
- Acceptance: Local fixture proves navigation, JavaScript-rendered content, timeout, and Response conversion.
- Done signal: Real browser adapter tests pass or are skipped only with documented local dependency gating.
- Source refs: `references/Scrapling/scrapling/fetchers/chrome.py`, `references/Scrapling/scrapling/engines/_browsers/_controllers.py`, `references/Scrapling/docs/fetching/dynamic.md`

## 4. Allowed domains and offsite filtering

- Phase: `phase-4-spider / spider-core`
- Priority: `P1`
- Owner: `spider`
- Size: `medium`
- Contract status: `draft`
- Contract: Add allowed_domains-style filtering and offsite request statistics before live crawling controls are layered on top.
- Ready when: Spider core scheduler and stats are stable.
- Write scope: `spider/`, `testdata/spider/`
- Test commands: `go test ./... -run TestSpiderAllowedDomains -count=1`
- Acceptance: Fake requests prove offsite drops, allowed host matching, and stats increments.
- Done signal: Allowed-domain tests pass.
- Source refs: `references/Scrapling/scrapling/spiders/engine.py`, `references/Scrapling/docs/spiders/advanced.md`

## 5. Crawler engine concurrency, domain limits, and download delay

- Phase: `phase-4-spider / spider-core`
- Priority: `P1`
- Owner: `spider`
- Size: `medium`
- Contract status: `draft`
- Contract: Add the crawler engine loop controls for global concurrency, per-domain concurrency, backpressure, context cancellation, and download delays.
- Ready when: Spider core request/result/session/scheduler contracts are validated.
- Write scope: `spider/`, `testdata/spider/`
- Test commands: `go test ./... -run TestSpiderEngineConcurrency -count=1`
- Acceptance: Fake sessions prove bounded concurrency, per-domain limits, cancellation, and deterministic delay behavior.
- Done signal: Spider engine concurrency tests pass.
- Source refs: `references/Scrapling/scrapling/spiders/engine.py`, `references/Scrapling/docs/spiders/architecture.md`, `references/Scrapling/docs/spiders/advanced.md`

## 6. Development response cache

- Phase: `phase-4-spider / spider-core`
- Priority: `P1`
- Owner: `spider`
- Size: `medium`
- Contract status: `draft`
- Contract: Add spider development cache behavior keyed by request fingerprint with encoded response bodies and cache hit/miss stats.
- Ready when: Request fingerprints and Response body helpers are validated.
- Write scope: `spider/cache.go`, `spider/cache_test.go`, `testdata/spider/cache/`
- Test commands: `go test ./... -run TestSpiderResponseCache -count=1`
- Acceptance: Temp-dir fixtures prove cache put/get/clear, binary-safe bodies, method separation, and cache stats.
- Done signal: Spider response cache tests pass.
- Source refs: `references/Scrapling/scrapling/spiders/cache.py`, `references/Scrapling/docs/spiders/advanced.md`

## 7. Gormes web-search tool adapter

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P2`
- Owner: `integration`
- Size: `medium`
- Contract status: `draft`
- Contract: Expose goscrapling as a production-friendly Go web search/scraping tool for Gormes without importing Gormes runtime dependencies into the core library.
- Ready when: Static fetcher and response objects are validated, Tool input/output schema is designed separately from library APIs, Gormes integration boundary is documented as an adapter, not a core dependency.
- Not ready when: The row tries to replace Gormes web search, browser tools, or channel rendering in one slice, The adapter needs live network access to prove its first behavior.
- Write scope: `integrations/gormes/`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md`
- Test commands: `go test ./... -run TestGormesIntegration -count=1`
- Acceptance: A fake tool call can fetch a local page, extract selected content, and return structured evidence, The adapter output includes enough evidence for Gormes to render or truncate without depending on goscrapling internals.
- Done signal: Integration tests pass without reaching the network, The Gormes boundary docs link back to the goscrapling strategy page.
- Source refs: `README.md`, `docs/research/go-scraping-oss-survey.md`, `docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md`

## 8. CLI interactive shell command surface

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `cli`
- Size: `large`
- Contract status: `draft`
- Contract: Map Scrapling's interactive shell concepts into a Go command surface with scripted evaluation first, before any full REPL dependency.
- Ready when: Static CLI extract behavior is stable.
- Not ready when: The row tries to build a full interactive REPL before scripted command fixtures exist.
- Write scope: `cmd/goscrapling/`, `internal/cli/`, `internal/cli/testdata/`
- Test commands: `go test ./... -run TestCLIShell -count=1`
- Acceptance: Scripted shell fixtures prove command evaluation and page shortcut behavior without live web access.
- Done signal: Shell command tests pass from local fixtures.
- Source refs: `references/Scrapling/scrapling/cli.py`, `references/Scrapling/docs/cli/interactive-shell.md`, `references/Scrapling/scrapling/core/shell.py`, `references/Scrapling/scrapling/core/_shell_signatures.py`

## 9. Public docs, examples, and API reference parity

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `docs`
- Size: `medium`
- Contract status: `draft`
- Contract: Maintain Go docs/examples for parser, adaptive scraping, fetchers, browser fetching, spiders, CLI, MCP, and migration guidance mapped from upstream docs.
- Ready when: Core APIs have stable examples for each subsystem.
- Write scope: `README.md`, `docs/`, `example_test.go`
- Test commands: `go test ./... -run TestExamples -count=1`
- Acceptance: Examples compile and docs point each major upstream feature group to a Go status or owned exclusion.
- Done signal: Example tests and docs checks pass.
- Source refs: `references/Scrapling/docs/index.md`, `references/Scrapling/docs/overview.md`, `references/Scrapling/docs/tutorials/migrating_from_beautifulsoup.md`, `references/Scrapling/docs/tutorials/replacing_ai.md`
<!-- PROGRESS:END -->
