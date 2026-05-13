# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Response metadata and selector contract

- Phase: `phase-1-response-fetcher / response`
- Priority: `P0`
- Owner: `fetcher`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Expose a Response that carries URL, status code, headers, final body URL, request metadata, and selector behavior over the parsed document.
- Ready when: A local response fixture can be constructed without network access.
- Not ready when: The row tries to implement JSON helpers, body decoding, or HTTP fetching in the same slice.
- Write scope: `response.go`, `response_test.go`, `document.go`
- Test commands: `go test ./... -run TestResponseMetadataAndSelectorContract -count=1`
- Acceptance: A local fixture can build a Response, inspect URL/status/headers, and query it with CSS.
- Done signal: Response metadata and selector contract test passes.
- Source refs: `references/Scrapling/docs/api-reference/response.md`, `references/Scrapling/docs/fetching/choosing.md`, `references/Scrapling/scrapling/fetchers/requests.py`, `references/Scrapling/scrapling/engines/static.py`

## 2. File-backed adaptive store with compatibility migration

- Phase: `phase-2-storage / persistent-store`
- Priority: `P1`
- Owner: `storage`
- Size: `medium`
- Contract status: `draft`
- Contract: Persist adaptive fingerprints beyond process memory with explicit schema versioning and deterministic migration tests.
- Ready when: The in-memory store contract remains stable after response/fetcher design.
- Write scope: `store.go`, `store_file.go`, `store_file_test.go`, `testdata/adaptive_store/`
- Test commands: `go test ./... -run TestFileStore -count=1`
- Acceptance: A file-backed store can save, reload, isolate domains, and reject incompatible schema versions visibly.
- Done signal: File store tests pass with temp directories only.
- Source refs: `references/Scrapling/docs/development/adaptive_storage_system.md`, `references/Scrapling/scrapling/core/storage.py`, `store.go`

## 3. BrowserFetcher interface and chromedp/playwright adapter decision

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `draft`
- Contract: Define a Go browser fetcher contract for dynamic pages, page actions, wait conditions, resource blocking, and response extraction before binding to an engine.
- Ready when: Static Response and FetcherSession behavior is validated.
- Not ready when: The row tries to implement stealth, proxy rotation, and browser actions in one slice.
- Write scope: `browser.go`, `browser_test.go`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- Test commands: `go test ./... -run TestBrowserFetcherContract -count=1`
- Acceptance: A fake browser engine fixture proves the public contract before choosing the real engine dependency.
- Done signal: Browser fetcher contract tests pass without launching a real browser.
- Source refs: `references/Scrapling/docs/fetching/dynamic.md`, `references/Scrapling/docs/fetching/stealthy.md`, `references/Scrapling/scrapling/fetchers/chrome.py`, `references/Scrapling/scrapling/fetchers/stealth_chrome.py`, `references/Scrapling/scrapling/engines/_browsers/_page.py`

## 4. Spider request, result, scheduler, and session contracts

- Phase: `phase-4-spider / spider-core`
- Priority: `P1`
- Owner: `spider`
- Size: `large`
- Contract status: `draft`
- Contract: Port Scrapling spider concepts into typed Go request/result/session/scheduler interfaces with deterministic fixture scheduling.
- Ready when: FetcherSession is complete and request/response behavior is fixture-backed.
- Not ready when: The row includes robots, cache, checkpoint, sitemap templates, and live crawling in the same slice.
- Write scope: `spider/`, `spider_test.go`, `testdata/spider/`
- Test commands: `go test ./... -run TestSpider -count=1`
- Acceptance: A fake fetcher drives deterministic scheduling, callback results, and session reuse.
- Done signal: Spider core tests pass using only fake fetchers and local fixtures.
- Source refs: `references/Scrapling/docs/spiders/architecture.md`, `references/Scrapling/docs/spiders/requests-responses.md`, `references/Scrapling/scrapling/spiders/request.py`, `references/Scrapling/scrapling/spiders/result.py`, `references/Scrapling/scrapling/spiders/scheduler.py`, `references/Scrapling/scrapling/spiders/session.py`

## 5. CLI extraction workflow parity

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P2`
- Owner: `cli`
- Size: `medium`
- Contract status: `draft`
- Contract: Map Scrapling CLI extract and shell workflows into a Go command surface only after core fetcher and parser behavior is stable.
- Ready when: Response and static fetcher APIs are stable enough for command output fixtures.
- Write scope: `cmd/goscrapling/`, `internal/cli/`, `testdata/cli/`
- Test commands: `go test ./... -run TestCLI -count=1`
- Acceptance: CLI fixtures prove deterministic extraction output and error messages without live network calls.
- Done signal: CLI tests pass from local fixtures.
- Source refs: `references/Scrapling/scrapling/cli.py`, `references/Scrapling/docs/cli/overview.md`, `references/Scrapling/docs/cli/extract-commands.md`, `references/Scrapling/docs/cli/interactive-shell.md`

## 6. Gormes web-search tool adapter

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P2`
- Owner: `integration`
- Size: `medium`
- Contract status: `draft`
- Contract: Expose goscrapling as a production-friendly Go web search/scraping tool for Gormes without importing Gormes runtime dependencies into the core library.
- Ready when: Static fetcher and response objects are validated., Tool input/output schema is designed separately from library APIs.
- Write scope: `integrations/gormes/`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- Test commands: `go test ./... -run TestGormesIntegration -count=1`
- Acceptance: A fake tool call can fetch a local page, extract selected content, and return structured evidence.
- Done signal: Integration tests pass without reaching the network.
- Source refs: `README.md`, `docs/research/go-scraping-oss-survey.md`
<!-- PROGRESS:END -->
