# Cross-Layer E2E Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add hermetic cross-layer E2E coverage that proves implemented goscrapling layers compose through public APIs and the CLI binary.

**Architecture:** Add one builder-ready `progress.json` row, then add one focused Go E2E test in `cmd/goscrapling`. The test uses local HTTP fixtures, temp storage, fake browser engine, and existing binary helpers; no live web, credentials, or new runtime features.

**Tech Stack:** Go `testing`, `httptest`, goscrapling public APIs, `spiders`, `integrations/gormes`, temp directories, generated progress docs.

---

### Task 1: Add the progress-ledger row

**Files:**
- Modify: `docs/content/building-goscrapling/architecture_plan/progress.json`
- Generated after validation: `docs/content/building-goscrapling/builder-loop/agent-queue.md`
- Generated after validation: `docs/content/building-goscrapling/builder-loop/next-slices.md`

- [ ] **Step 1: Insert the row under `phase-5-cli-tooling/tool-surfaces`**

Add this item after `Full CLI E2E validation harness`:

```json
{
  "name": "Cross-layer local E2E validation harness",
  "status": "planned",
  "priority": "P1",
  "contract": "Prove the implemented local layers compose end to end: static fetcher, response selectors, parser adaptive storage, spider flow, CLI binary extraction, and Gormes browser-tool adapter using only hermetic fixtures.",
  "contract_status": "fixture_ready",
  "slice_size": "medium",
  "execution_owner": "integration",
  "source_refs": [
    "cmd/goscrapling/full_e2e_test.go",
    "cmd/goscrapling/main_test.go",
    "fetchers/fetcher_test.go",
    "parser/adaptive_selector_test.go",
    "spiders/spider_test.go",
    "integrations/gormes/browser_tools_test.go",
    "docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md"
  ],
  "ready_when": [
    "Existing layer-specific tests are present and the harness can use local fixtures only."
  ],
  "not_ready_when": [
    "The proof requires live web access, credentials, a real browser, stealth behavior, proxy rotation, or unplanned production feature work."
  ],
  "write_scope": [
    "cmd/goscrapling/cross_layer_e2e_test.go",
    "docs/content/building-goscrapling/architecture_plan/progress.json",
    "docs/content/building-goscrapling/builder-loop/agent-queue.md",
    "docs/content/building-goscrapling/builder-loop/next-slices.md"
  ],
  "test_commands": [
    "go test ./cmd/goscrapling -run TestGoscraplingCrossLayerLocalEndToEnd -count=1",
    "go test ./... -count=1",
    "go run ./cmd/progress validate",
    "jq empty docs/content/building-goscrapling/architecture_plan/progress.json",
    "git diff --check"
  ],
  "acceptance": [
    "A local server fixture is fetched and parsed through public goscrapling APIs.",
    "Adaptive selector save and relocate uses a temp persistent store.",
    "A spider flow follows a local link and emits an item from the detail page.",
    "The goscrapling CLI binary writes selected fixture output to disk.",
    "The Gormes browser extraction adapter returns rendered markdown from a fake browser engine."
  ],
  "done_signal": [
    "TestGoscraplingCrossLayerLocalEndToEnd and the full completion gate pass without live network dependencies."
  ]
}
```

- [ ] **Step 2: Validate the planner row**

Run:

```sh
go run ./cmd/progress validate
```

Expected: command exits 0.

- [ ] **Step 3: Regenerate generated queue docs**

Run:

```sh
go run ./cmd/progress write
```

Expected: generated builder-loop markdown updates cleanly.

### Task 2: Add the failing cross-layer E2E test

**Files:**
- Create: `cmd/goscrapling/cross_layer_e2e_test.go`

- [ ] **Step 1: Write the test**

Create `cmd/goscrapling/cross_layer_e2e_test.go` with a single test named `TestGoscraplingCrossLayerLocalEndToEnd`. It should use `httptest.Server`, `goscrapling.NewFileStore(t.TempDir())`, `goscrapling.Fetcher`, `goscrapling.Parse`, `spiders.Crawler`, `buildGoscraplingBinary`, and `gormes.BrowserExtractionAdapter` with a fake `browser.BrowserEngine`.

- [ ] **Step 2: Run the focused test to verify behavior**

Run:

```sh
go test ./cmd/goscrapling -run TestGoscraplingCrossLayerLocalEndToEnd -count=1
```

Expected before any implementation fix: if the test exposes a missing composition edge, it fails with the exact missing behavior. If it passes immediately, keep it as coverage-only evidence and do not add production code.

### Task 3: Complete row and validate

**Files:**
- Modify: `docs/content/building-goscrapling/architecture_plan/progress.json`
- Generated: `docs/content/building-goscrapling/builder-loop/agent-queue.md`
- Generated: `docs/content/building-goscrapling/builder-loop/next-slices.md`

- [ ] **Step 1: If focused E2E passes, mark row complete**

Set:

```json
"status": "complete",
"contract_status": "validated"
```

Update `done_signal` to include the exact passing focused command.

- [ ] **Step 2: Run completion gate**

Run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress write
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Expected: all commands exit 0.

- [ ] **Step 3: Commit and push**

Run:

```sh
git add docs/superpowers/specs/validation/2026-05-20-cross-layer-e2e-design.md docs/superpowers/plans/2026-05-20-cross-layer-e2e.md docs/content/building-goscrapling/architecture_plan/progress.json docs/content/building-goscrapling/builder-loop/agent-queue.md docs/content/building-goscrapling/builder-loop/next-slices.md cmd/goscrapling/cross_layer_e2e_test.go
git commit -m "test: add cross-layer e2e harness"
git push origin main
```

Expected: commit succeeds and push reports success.
