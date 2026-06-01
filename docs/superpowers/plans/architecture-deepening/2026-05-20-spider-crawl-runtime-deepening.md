# Spider Crawl Runtime Deepening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deepen the goscrapling spider crawl engine by moving `Crawler.Run` orchestration state into a private crawl runtime without changing public behavior.

**Architecture:** Keep public `spiders.Crawler` and all public request/session/result types unchanged. Add a private `crawlRuntime` module that owns scheduler loop state, result mutation, active task accounting, allowed-domain checks, domain limiters, and sleep policy, matching Scrapling's deep engine seam while staying Go-native.

**Tech Stack:** Go, package-private structs/functions, existing spider tests as characterization tests, repository validation gates.

---

### Task 1: Characterize current spider behavior

**Files:**
- Test: `spiders/spider_test.go`
- Test: `spiders/allowed_domains_test.go`
- Test: `spiders/engine_concurrency_test.go`

- [ ] **Step 1: Run focused spider tests before editing**

Run:

```sh
go test ./spiders -count=1
```

Expected: PASS. This is the characterization baseline for the no-behavior refactor.

### Task 2: Introduce private crawl runtime

**Files:**
- Create: `spiders/crawl_runtime.go`
- Modify: `spiders/crawler.go`

- [ ] **Step 1: Create `crawlRuntime`**

Move the orchestration state currently local to `Crawler.Run` into a private struct with fields for the source `Crawler`, context/cancel, scheduler, allowed domains, concurrent request limit, domain limiters, sleep function, result, active task count, and stop error.

- [ ] **Step 2: Move the run loop behind `crawlRuntime.run`**

Keep `Crawler.Run(ctx, start)` as a thin constructor/validator that builds `crawlRuntime` and calls `runtime.run(start)`. Preserve exact error behavior for nil context and missing sessions.

- [ ] **Step 3: Move result mutation into runtime methods**

Extract package-private methods such as `enqueueStartRequests`, `startReadyTasks`, `handleTaskResult`, `handleOutput`, and `doneChannel`. These methods must preserve existing stats and cancellation behavior.

### Task 3: Validate and deliver

**Files:**
- Modified: `spiders/crawler.go`
- Created: `spiders/crawl_runtime.go`
- Created: `docs/superpowers/plans/architecture-deepening/2026-05-20-spider-crawl-runtime-deepening.md`

- [ ] **Step 1: Run focused tests**

Run:

```sh
go test ./spiders -count=1
```

Expected: PASS.

- [ ] **Step 2: Run full gate**

Run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Expected: all commands exit 0.

- [ ] **Step 3: Commit and push**

Run:

```sh
git add spiders/crawler.go spiders/crawl_runtime.go docs/superpowers/plans/architecture-deepening/2026-05-20-spider-crawl-runtime-deepening.md
git commit -m "refactor: deepen spider crawl runtime"
git push origin main
```
