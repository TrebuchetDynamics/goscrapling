# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Shell browser shortcut boundary

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `cli`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Expose explicit scripted-shell boundaries for Scrapling's fetch(url) and stealthy_fetch(url) browser shortcuts so goscrapling reports the unsupported shell shortcut without launching a browser or suggesting stealth bypass behavior.
- Ready when: Static shell shortcuts and view boundary are validated., The fixture asserts command errors only and does not start a browser.
- Not ready when: The implementation would launch a real browser, add stealth behavior, introduce browser session state, or expand shell parsing beyond explicit fetch/stealthy_fetch boundaries.
- Write scope: `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`, `docs/research/scrapling-parity-matrix.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`, `docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md`
- Test commands: `go test ./internal/cli/shell -run TestCLIShellBrowserShortcutBoundary -count=1`, `go test ./internal/cli/shell -count=1`, `go test ./... -count=1`
- Acceptance: fetch(url) in shell -c returns a parse error that names the unsupported browser shell shortcut and points to goscrapling extract fetch., stealthy_fetch(url) in shell -c returns a parse error that names the unsupported stealth browser shell shortcut and points to goscrapling extract stealthy-fetch., Neither boundary path launches a browser, performs network I/O, or updates page history.
- Done signal: TestCLIShellBrowserShortcutBoundary passes from command-output fixtures.
- Source refs: `references/Scrapling/scrapling/core/shell.py`, `references/Scrapling/scrapling/core/_shell_signatures.py`, `references/Scrapling/docs/cli/interactive-shell.md`, `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`
<!-- PROGRESS:END -->
