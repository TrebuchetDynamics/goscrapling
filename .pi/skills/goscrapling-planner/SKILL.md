---
name: goscrapling-planner
description: Use when turning goscrapling parity gaps, broad requests, roadmap items, feature-map changes, or upstream Scrapling behavior into one builder-ready progress.json row.
---

# goscrapling-planner

## Purpose

Convert a known Scrapling parity gap into one builder-sized row that another
agent can implement without rediscovering upstream behavior or guessing scope.

## Inputs

- User goal and selected upstream behavior.
- `docs/content/building-goscrapling/architecture_plan/progress.json`.
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`.
- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`.
- Scrapling source refs under `references/Scrapling` when present.
- Existing goscrapling tests and public APIs.

## Builder-Ready Row Contract

A row is builder-ready only when it has:

- `name`, `status`, `priority`, `contract`, `contract_status`, `slice_size`,
  and `execution_owner`;
- exact `source_refs` pointing to upstream files/docs/tests;
- `ready_when` and, for broad or risky work, `not_ready_when`;
- narrow `write_scope`;
- `test_commands` or an explicit `no_test_required` reason;
- observable `acceptance` and `done_signal`.

## Planning Rules

- Prefer one vertical slice that proves behavior end to end.
- Split umbrella rows before handing work to a builder.
- Preserve Go-native API design, but cite the Scrapling-visible behavior being
  preserved or the owned divergence being chosen.
- Keep tests hermetic: `httptest`, temp dirs, fake browser engines, and local
  fixtures before live network or browser tests.
- Update the feature map and coverage ledger when a new upstream source class
  becomes relevant.

## Validation

After editing progress or planning docs, run:

```sh
go run ./cmd/progress validate
go run ./cmd/progress write
git diff --check
```

Run `go test ./... -count=1` when the edit changes generated docs, schema,
progress rows, or tests that may be checked by repository test code.

## Done

A planner pass is done when one row is builder-ready, linked docs agree with
that row, and validation output is reported exactly.
