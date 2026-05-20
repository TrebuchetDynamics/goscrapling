# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Gormes browser extraction tools

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P1`
- Owner: `integration`
- Size: `medium`
- Contract status: `fixture_ready`
- Contract: Expose deterministic Gormes-facing browser extraction tool adapters for rendered markdown, links, semantic tree, structured data, and basic form/interactive-element discovery using fake browser engines and tool-call fixtures.
- Ready when: Browser markdown and semantic tree rows are validated., A fake browser engine can return deterministic rendered HTML for adapter tests.
- Not ready when: The row requires live LLM calls, live web access, a real browser, or Gormes runtime changes outside the adapter boundary., The row attempts a full MCP server implementation instead of Gormes-facing extraction adapters.
- Write scope: `integrations/gormes/`, `integrations/gormes/testdata/`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md`, `docs/content/building-goscrapling/architecture_plan/progress.json`
- Test commands: `go test ./... -run TestGormesBrowserExtractionTools -count=1`
- Acceptance: Fake tool-call fixtures prove rendered markdown, links, semantic tree, structured data, and interactive/form discovery responses., Tool outputs are deterministic, text-first, and do not require live LLM, live browser, or live network access., The adapter names and schemas are documented as goscrapling-owned while inspired by Lightpanda MCP tool taxonomy.
- Done signal: TestGormesBrowserExtractionTools passes and the integration docs identify the fake-engine proof path.
- Source refs: `references/Scrapling/scrapling/core/ai.py`, `references/Scrapling/docs/ai/mcp-server.md`, `references/Scrapling/docs/api-reference/mcp-server.md`, `external/lightpanda-io/browser@5905319e78541110a0b4065b07ec7ce53f93a660:src/mcp/tools.zig`, `docs/content/building-goscrapling/strategy/portfolio-and-gormes-fit.md`

## 2. Development response cache

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

## 3. Gormes web-search tool adapter

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

## 4. CLI interactive shell command surface

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

## 5. Public docs, examples, and API reference parity

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
