# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Fetch safety controls: robots and private network blocking

- Phase: `phase-1-response-fetcher / static-fetcher`
- Priority: `P1`
- Owner: `fetcher`
- Size: `medium`
- Contract status: `fixture_ready`
- Contract: Add explicit operator-visible fetch safety controls for robots.txt enforcement and private-network/CIDR blocking across static fetch paths, with browser fetch integration planned as a follow-up rather than an implicit side effect.
- Ready when: Static fetcher request options and error taxonomy are validated., Hermetic httptest fixtures can serve target pages and robots.txt responses.
- Not ready when: The row attempts crawler politeness, rate limiting, or browser resource interception beyond robots and private network blocking., The row requires live DNS, live web access, or privileged network configuration.
- Write scope: `fetchers/fetcher_safety.go`, `fetchers/fetcher_safety_test.go`, `fetchers/fetcher_options.go`, `fetchers/fetcher_errors.go`, `docs/content/building-goscrapling/architecture_plan/boundaries.md`, `docs/content/building-goscrapling/architecture_plan/progress.json`
- Test commands: `go test ./... -run TestFetcherSafetyControls -count=1`
- Acceptance: httptest fixtures prove robots.txt allow and disallow decisions without live network access., Private IPv4, IPv6, localhost, and configured CIDR targets are rejected before request dispatch when blocking is enabled., Safety decisions return typed/operator-visible errors and remain opt-in or explicitly configured.
- Done signal: TestFetcherSafetyControls passes and boundaries.md documents the safety-control contract.
- Source refs: `references/Scrapling/docs/fetching/static.md`, `references/Scrapling/docs/fetching/dynamic.md`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/network/layer/RobotsLayer.zig`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/network/Robots.zig`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/network/IpFilter.zig`, `fetchers/`

## 2. Browser fetch markdown dump

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `fixture_ready`
- Contract: Add a browser-rendered markdown dump path that turns rendered HTML into compact AI-readable Markdown while preserving headings and links and stripping navigation, script, style, and UI noise by default.
- Ready when: BrowserFetcher interface and fake browser engine contract are validated., A local fixture can exercise rendered HTML without launching a real browser.
- Not ready when: The implementation requires live web access or a real browser process for the core test., The row expands into CLI flags, MCP tools, or full readability extraction beyond markdown dump behavior.
- Write scope: `engines/browser/browser_markdown.go`, `engines/browser/browser_markdown_test.go`, `testdata/browser/`, `docs/content/building-goscrapling/architecture_plan/progress.json`
- Test commands: `go test ./... -run TestBrowserFetchMarkdownDump -count=1`
- Acceptance: A fake browser engine returns HTML containing headings, links, nav, script, style, and article content., The markdown dump preserves headings and links in deterministic text output., The markdown dump strips script/style content and default navigation or UI noise without live network access.
- Done signal: TestBrowserFetchMarkdownDump passes and the row is marked complete with validated evidence.
- Source refs: `references/Scrapling/docs/fetching/dynamic.md`, `references/Scrapling/scrapling/fetchers/chrome.py`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:README.md --dump markdown`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/browser/markdown.zig`, `engines/browser/browser.go`, `engines/browser/browser_test.go`

## 3. Browser semantic tree extraction

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `fixture_ready`
- Contract: Expose a browser-rendered semantic tree for AI agents that summarizes visible elements with stable node IDs, role, accessible name/text, tag name, XPath or selector path, value state, disabled state, and interactivity.
- Ready when: Browser markdown or rendered HTML fixture path can provide deterministic HTML to the extractor., The row remains limited to semantic extraction and does not require browser actions.
- Not ready when: The implementation attempts full accessibility tree parity or CDP backend node integration., The implementation requires live browser execution for the core test.
- Write scope: `engines/browser/browser_semantic.go`, `engines/browser/browser_semantic_test.go`, `parser/`, `testdata/browser/`, `docs/content/building-goscrapling/architecture_plan/progress.json`
- Test commands: `go test ./... -run TestBrowserSemanticTreeExtraction -count=1`
- Acceptance: A local HTML fixture with buttons, links, inputs, labels, disabled controls, and nested content returns a deterministic semantic tree., Each relevant element includes role, name or text, tag, XPath or selector path, interactivity, and value/disabled state where applicable., Noise nodes such as script/style and whitespace-only text are omitted.
- Done signal: TestBrowserSemanticTreeExtraction passes and documents the owned Go-native semantic tree contract.
- Source refs: `references/Scrapling/docs/fetching/dynamic.md`, `references/Scrapling/scrapling/fetchers/chrome.py`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/SemanticTree.zig`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/mcp/tools.zig semantic_tree`, `engines/browser/browser.go`, `parser/`

## 4. Development response cache

- Phase: `phase-4-spider / spider-core`
- Priority: `P1`
- Owner: `spider`
- Size: `medium`
- Contract status: `draft`
- Contract: Add spider development cache behavior keyed by request fingerprint with encoded response bodies and cache hit/miss stats.
- Ready when: Request fingerprints and Response body helpers are validated.
- Write scope: `spiders/cache.go`, `spiders/cache_test.go`, `testdata/spiders/cache/`
- Test commands: `go test ./... -run TestSpiderResponseCache -count=1`
- Acceptance: Temp-dir fixtures prove cache put/get/clear, binary-safe bodies, method separation, and cache stats.
- Done signal: Spider response cache tests pass.
- Source refs: `references/Scrapling/scrapling/spiders/cache.py`, `references/Scrapling/docs/spiders/advanced.md`

## 5. Gormes web-search tool adapter

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

## 6. CLI interactive shell command surface

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

## 7. Public docs, examples, and API reference parity

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
