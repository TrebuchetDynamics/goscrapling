# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Shell scripted help and namespace listing

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P3`
- Owner: `cli`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Expose a bounded goscrapling shell help() shortcut that mirrors Scrapling's shell banner intent for the scripted Go shell by listing supported shortcuts, page variables, curl helpers, and the no-REPL boundary without touching the network.
- Ready when: The shell remains scripted with -c; no interactive REPL is required., The help fixture asserts command output only and does not fetch any URL.
- Not ready when: The implementation attempts a full interactive REPL, IPython-compatible introspection, browser helpers, or live network examples.
- Write scope: `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`, `docs/research/scrapling-parity-matrix.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`, `docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md`
- Test commands: `go test ./internal/cli/shell -run TestCLIShellHelpShortcut -count=1`, `go test ./internal/cli/shell -count=1`, `go test ./... -count=1`
- Acceptance: help() in shell -c prints the supported scripted shortcuts and variables without requiring a fetched page., The output lists get, post, put, delete, page, response, pages, uncurl, and curl2fetcher., The output explicitly says interactive REPL support is not implemented in goscrapling.
- Done signal: TestCLIShellHelpShortcut passes from command-output fixtures.
- Source refs: `references/Scrapling/scrapling/core/shell.py`, `references/Scrapling/docs/cli/interactive-shell.md`, `internal/cli/shell/shell_command.go`, `internal/cli/shell/shell_integration_test.go`
<!-- PROGRESS:END -->
