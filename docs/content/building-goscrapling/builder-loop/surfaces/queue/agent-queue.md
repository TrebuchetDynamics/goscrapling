# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Gormes provider failure diagnostics

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P2`
- Owner: `integration`
- Size: `small`
- Contract status: `fixture_ready`
- Contract: Add integration-local provider failure diagnostics for Gormes static providers, inspired by SearXNG processor error handling and suspended-engine reporting: record unresponsive providers with error type, optional suspended flag/reason, and deterministic diagnostic snapshots without adding retries, scheduler behavior, provider fanout, or live federation.
- Ready when: Gormes result container merge/scoring is validated with local fixtures., The builder scopes behavior to recording and surfacing diagnostics from fake/local provider results, not retrying or scheduling provider calls.
- Not ready when: The row tries to add live provider federation, retry/backoff scheduling, provider suspension timers, query mini-syntax, plugin hooks, config-file loading, browser providers, or runtime Gormes registration., The implementation copies SearXNG AGPL source instead of reimplementing the diagnostics pattern from the row contract., The tests require live web access, real external providers, browser launches, timers that sleep, or credentials.
- Write scope: `integrations/gormes/statictools/`, `integrations/gormes/static_tools.go`, `integrations/gormes/testdata/`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`, `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`, `docs/research/scrapling-parity-matrix.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`
- Test commands: `go test ./integrations/gormes/statictools -run TestGormesProviderFailureDiagnostics -count=1`, `go test ./integrations/gormes/... -count=1`, `go test ./... -count=1`
- Acceptance: Local fixtures can record provider diagnostics with provider name, error type, message, suspended flag, and suspension reason without contacting the network., Diagnostics appear in result-container snapshots alongside merged results and remain stable in insertion order for deterministic tool output., A provider call failure can be converted into a diagnostic entry without retrying, sleeping, scheduling another provider, or hiding the operator-visible error., The adapter remains integration-local and explicitly excludes retry/backoff timers, live federation, query syntax, plugin hooks, browser providers, config loading, and runtime Gormes registration.
- Done signal: TestGormesProviderFailureDiagnostics passes from local fixtures., Full repository validation passes after generated queue docs list the next remaining row.
- Source refs: `external/searxng/searxng@37187dc2d:searx/search/processors/abstract.py`, `external/searxng/searxng@37187dc2d:searx/results.py`, `external/searxng/searxng@37187dc2d:searx/webutils.py`, `integrations/gormes/statictools/providers.go`, `integrations/gormes/statictools/results.go`, `integrations/gormes/statictools/static_tools_test.go`

## 2. Shell browser shortcut boundary

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
