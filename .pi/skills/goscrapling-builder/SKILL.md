---
name: goscrapling-builder
description: Use when implementing exactly one builder-ready goscrapling progress.json row after a failing test exists and the row has source refs, write scope, acceptance, and test commands.
---

# goscrapling-builder

## Purpose

Implement exactly one builder-ready row with the minimum Go code needed to pass
its tests while preserving Scrapling-visible behavior and documented
Go-native boundaries.

## Required Inputs

- One `progress.json` row that satisfies the planner row contract.
- The failing test produced by `goscrapling-tdd-slice`.
- All row `source_refs`, including Scrapling refs and any external design refs.
- Existing package patterns in the row's `write_scope`.
- `docs/content/building-goscrapling/architecture_plan/boundaries.md` when the
  row touches safety, stealth, browser behavior, licensing, or external refs.

## Row Intake Checklist

Before implementation, confirm and report:

- row `status` is `planned` or `in_progress`, not `complete`;
- row `slice_size` is not `umbrella`;
- row has `source_refs`, `write_scope`, `test_commands`, `acceptance`, and
  `done_signal`;
- the focused failing test fails for the expected missing behavior;
- every edited file is inside `write_scope`, or the row is returned to planner.

## Source Reference Rules

- Scrapling refs define parity behavior.
- External refs such as Lightpanda are design references only unless the row
  says otherwise.
- Do not copy AGPL or incompatible source into goscrapling; reimplement the
  behavior from the row contract and cite the reference in docs/tests.
- If a source ref is unavailable locally, use existing ledgers/docs and report
  the missing checkout; do not invent exact behavior.

## Implementation Rules

- Touch only the row write scope unless the failing test proves the scope is
  wrong; return to planner if scope expands across multiple subsystems.
- Prefer standard-library Go and existing repository patterns.
- Keep APIs small and stable before adding convenience layers.
- Keep core tests hermetic: no live network, live browser, live LLM, or
  credential dependency unless the row explicitly marks an optional smoke test.
- Do not implement browser stealth, proxy rotation, robots, private-network
  controls, or crawling behavior as incidental side effects of another row.
- Do not mark parity from prose alone; covered behavior needs tests.

## Red-Green-Refactor Loop

1. Run the focused failing test and confirm the expected failure.
2. Implement the smallest code path that satisfies the row acceptance.
3. Run the focused test until green.
4. Refactor only inside the row scope while the focused test remains green.
5. Run the full completion gate.
6. Update `progress.json` only after validation evidence exists.

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
- ensure acceptance and done signal describe the exact evidence that passed;
- run `go run ./cmd/progress write` if generated builder-loop docs changed.

## Stop And Return To Planner

Stop instead of coding when:

- the row is missing builder-ready metadata;
- the first failing test requires multiple rows or an umbrella scope;
- a source reference changes licensing or product boundary assumptions;
- implementation would require live services for the core proof;
- write scope must expand beyond the row.

## Done

A builder pass is done when the focused test and full gate pass, the row status
matches the evidence, generated docs are updated when needed, and the report
lists exact commands, outputs, files changed, commit hash, and push status.
