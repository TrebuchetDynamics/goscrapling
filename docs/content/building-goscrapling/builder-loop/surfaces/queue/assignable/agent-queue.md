# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Shell view command boundary

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `cli`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Map Scrapling's shell view(page) helper into a bounded Go-native scripted shell behavior that avoids launching a browser during tests while making the visualization boundary explicit.
- Ready when: The shell remains scripted with -c; no interactive REPL is required., The visualization fixture uses local httptest HTML and no live browser launch.
- Not ready when: The implementation would open a real browser, depend on OS desktop integration, require live network access, or expand into a full REPL visualization system.
- Write scope: `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`, `docs/research/scrapling-parity-matrix.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`, `docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md`
- Test commands: `go test ./internal/cli/shell -run TestCLIShellViewCommandBoundary -count=1`, `go test ./internal/cli/shell -count=1`, `go test ./... -count=1`
- Acceptance: view(page) after a scripted fetch writes or reports a deterministic local visualization artifact without opening a browser., view(response) is accepted as an alias for the current page., view(page) before any fetch returns a parse error without panics or filesystem side effects.
- Done signal: TestCLIShellViewCommandBoundary passes from local fixtures.
- Source refs: `references/Scrapling/scrapling/core/shell.py`, `references/Scrapling/docs/cli/interactive-shell.md`, `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`
<!-- PROGRESS:END -->
