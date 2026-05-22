# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Public docs, examples, and API reference parity

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
