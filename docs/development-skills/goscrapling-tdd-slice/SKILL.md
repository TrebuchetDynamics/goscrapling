---
name: goscrapling-tdd-slice
description: Use when adding or fixing goscrapling runtime behavior, before implementation code, to create the smallest failing test for one builder-ready progress.json row.
---

# goscrapling-tdd-slice

## Purpose

Prove the selected row's contract with a failing test before implementation.
Use this for runtime behavior, not docs-only skill or planning edits.

## Loop

1. Pick one builder-ready `progress.json` row.
2. Write the smallest test that captures the row contract.
3. Run the focused test and confirm it fails for the expected reason.
4. Hand the failing test to `goscrapling-builder` for implementation.
5. After implementation, run the focused test, then `go test ./... -count=1`.
6. Refactor only after tests are green.
7. Update `progress.json`, feature map, and coverage ledger if behavior status
   changed.

## Test Preferences

- Parser behavior: string fixtures.
- Fetcher behavior: `httptest.Server`.
- Store behavior: `t.TempDir()`.
- Browser behavior: fake browser engine first, real browser smoke tests later.
- Spider behavior: fake fetcher and deterministic scheduler fixtures.
- CLI behavior: command-output fixtures and local files.
- Integration behavior: fake tool calls; no live LLM dependency.

## Stop Conditions

Stop and return to planner if the row requires multiple subsystems, live
credentials, live web access, unresolved API design, browser stealth policy, or
proxy/crawling behavior not explicitly named in the row.

## Done

A TDD-slice pass is done when the failing focused test is committed to the row
scope, its failure reason is recorded, and the next builder command is clear.
