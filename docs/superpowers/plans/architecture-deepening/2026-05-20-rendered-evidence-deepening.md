# Rendered Evidence Deepening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Collapse duplicated browser/Gormes HTML evidence extraction behind one deep module without changing public behavior.

**Architecture:** Keep public browser and Gormes interfaces unchanged. Add a rendered evidence module in `engines/browser` that parses HTML once and exposes markdown, semantic nodes, links, structured data, and interactive nodes for existing callers.

**Tech Stack:** Go, `golang.org/x/net/html`, existing browser and Gormes tests as characterization tests.

---

### Task 1: Baseline

- [x] Run `go test ./engines/browser ./integrations/gormes -count=1` before editing.

### Task 2: Deepen rendered evidence

**Files:**
- Create: `engines/browser/rendered_evidence.go`
- Modify: `engines/browser/browser_markdown.go`
- Modify: `engines/browser/browser_semantic.go`
- Modify: `integrations/gormes/browser_tools.go`

Steps:
- [ ] Move shared HTML parse, visible text, attribute, skip, markdown, semantic, link, and structured-data extraction into `engines/browser`.
- [ ] Keep `HTMLToMarkdown` and `HTMLSemanticNodes` as compatibility wrappers.
- [ ] Update Gormes tools to call `browser.NewRenderedEvidence` instead of reparsing HTML locally.

### Task 3: Validate and deliver

- [ ] Run `go test ./engines/browser ./integrations/gormes -count=1`.
- [ ] Run `go test ./... -count=1`.
- [ ] Run `go run ./cmd/progress validate`.
- [ ] Run `jq empty docs/content/building-goscrapling/architecture_plan/progress.json`.
- [ ] Run `git diff --check`.
- [ ] Commit and push with `refactor: deepen rendered evidence extraction`.
