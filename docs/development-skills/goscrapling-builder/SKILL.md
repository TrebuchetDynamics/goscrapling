---
name: goscrapling-builder
description: Use when implementing exactly one builder-ready goscrapling progress.json row after a failing test exists and the row has source refs, write scope, acceptance, and test commands.
---

# goscrapling-builder

## Purpose

Implement exactly one builder-ready row with the minimum Go code needed to pass
its tests and preserve Scrapling-visible behavior.

## Inputs

- One `progress.json` row that satisfies the planner row contract.
- Failing test from `goscrapling-tdd-slice`.
- Source refs listed on the row.
- Existing package patterns in the row's `write_scope`.

## Implementation Rules

- Touch only the row write scope unless the failing test proves the scope is
  wrong; return to planner if scope expands across multiple subsystems.
- Prefer standard-library Go and existing repository patterns.
- Keep APIs small and stable before adding convenience layers.
- Do not add live network dependencies to core tests.
- Do not implement browser stealth, proxy rotation, or crawling controls as
  incidental side effects of another row.
- Do not mark parity from prose alone; covered behavior needs tests.

## Completion Gate

Before closing the row, run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Then update `progress.json`:

- set `status` to `complete`;
- set `contract_status` to `validated`;
- ensure acceptance and done signal describe the evidence that passed.

## Done

A builder pass is done when the focused test and full gate pass, the row status
matches the evidence, and the report lists exact commands, outputs, and files
changed.
