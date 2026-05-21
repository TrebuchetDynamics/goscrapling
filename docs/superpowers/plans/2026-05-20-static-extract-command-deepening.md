# Static Extract Command Deepening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deepen the static extract CLI implementation without changing command behavior, output, or parse errors.

**Architecture:** Keep `Run` as the public CLI entry point. Move extract-specific parse/plan/execute/render/write behavior into an internal extract command module.

**Tech Stack:** Go, existing `internal/cli` and `cmd/goscrapling` tests as characterization coverage.

---

### Task 1: Baseline

- [x] Run `go test ./internal/cli ./cmd/goscrapling -count=1` before editing.

### Task 2: Deepen extract command module

**Files:**
- Modify: `internal/cli/run.go`
- Create: `internal/cli/extract_command.go`

Steps:
- [ ] Keep `Run` as the top-level command router.
- [ ] Move extract-specific option parsing and request planning into `extract_command.go`.
- [ ] Keep static fetch, render, and write behavior internal to the extract command module.
- [ ] Preserve the existing `usage` string and `ErrParse` behavior.

### Task 3: Validate and deliver

- [ ] Run `go test ./internal/cli ./cmd/goscrapling -count=1`.
- [ ] Run `go test ./... -count=1`.
- [ ] Run `go run ./cmd/progress validate`.
- [ ] Run `jq empty docs/content/building-goscrapling/architecture_plan/progress.json`.
- [ ] Run `git diff --check`.
- [ ] Commit and push with `refactor: deepen static extract command`.
