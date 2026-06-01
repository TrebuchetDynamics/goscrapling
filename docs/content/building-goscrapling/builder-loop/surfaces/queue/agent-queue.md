# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Shell pages history expression shortcuts

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `cli`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Expose bounded scripted goscrapling shell expressions for inspecting pages history entries so the Go shell covers Scrapling's documented pages[0]/pages[-1] history exploration without implementing a full REPL.
- Ready when: The shell remains scripted with -c; no interactive REPL is required., History fixtures use local httptest pages only.
- Not ready when: The implementation attempts general Python indexing, arbitrary expressions, full REPL support, or live network examples.
- Write scope: `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`, `docs/research/scrapling-parity-matrix.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`, `docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md`
- Test commands: `go test ./internal/cli/shell -run TestCLIShellPagesHistoryExpressions -count=1`, `go test ./internal/cli/shell -count=1`, `go test ./... -count=1`
- Acceptance: pages[0].url and pages[-1].url return the oldest and newest retained response URLs from scripted shell history., pages[index].status returns the selected response status code., Out-of-range pages indexes return a parse error without panics.
- Done signal: TestCLIShellPagesHistoryExpressions passes from local httptest fixtures.
- Source refs: `references/Scrapling/scrapling/core/shell.py`, `references/Scrapling/docs/cli/interactive-shell.md`, `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`
<!-- PROGRESS:END -->
