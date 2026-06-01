# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Shell static method shortcuts beyond get

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `cli`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Extend the scripted goscrapling shell shortcuts from get(url) to the remaining upstream static Fetcher methods post, put, and delete with local request-option fixtures.
- Ready when: The shell remains scripted with -c; no interactive REPL is required., Static method shortcut tests use httptest fixtures only.
- Not ready when: The implementation would require browser fetchers, async sessions, full REPL support, or live network access.
- Write scope: `internal/cli/shell_command.go`, `internal/cli/shell_test.go`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`, `docs/research/scrapling-parity-matrix.md`, `docs/content/building-goscrapling/builder-loop/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/next-slices.md`, `docs/content/building-goscrapling/builder-loop/builder-loop-handoff.md`
- Test commands: `go test ./internal/cli -run TestCLIShellHTTPMethodShortcuts -count=1`, `go test ./internal/cli -count=1`, `go test ./... -count=1`
- Acceptance: post, put, and delete scripted shell shortcuts issue the matching HTTP methods to local httptest fixtures., Request body and header options for method shortcuts are parsed narrowly enough to mirror existing extract command fixtures., page, response, and pages shortcuts update after each method call.
- Done signal: TestCLIShellHTTPMethodShortcuts passes from local fixtures.
- Source refs: `references/Scrapling/scrapling/core/shell.py`, `references/Scrapling/scrapling/core/_shell_signatures.py`, `internal/cli/shell_command.go`, `internal/cli/shell_test.go`
<!-- PROGRESS:END -->
