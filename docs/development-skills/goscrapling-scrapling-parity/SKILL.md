---
name: goscrapling-scrapling-parity
description: Use when mapping upstream Scrapling behavior into goscrapling, auditing coverage, classifying parity status, or deciding whether a Go behavior is covered, partial, planned, vague, owned, excluded, or missing.
---

# goscrapling-scrapling-parity

## Purpose

Extract behavior atoms from upstream Scrapling and map them to goscrapling
code, tests, docs, and progress rows. This skill answers what parity means
before planner or builder work begins.

## Behavior Atom

Record each meaningful behavior with:

- upstream surface and exact source ref;
- trigger or API call;
- visible contract;
- state and side effects;
- local Go target;
- current progress row;
- validation command;
- risk if behavior drifts.

## Source Order

1. `references/Scrapling/scrapling/**`.
2. `references/Scrapling/docs/**`.
3. Existing goscrapling tests and docs.
4. `progress.json`.

If `references/Scrapling` is absent, use existing ledgers and report that the
upstream checkout was unavailable; do not invent source refs.

## Status Vocabulary

- `covered`: repository evidence and tests exist.
- `partial`: working Go behavior exists, but Scrapling-visible behavior is not
  fully covered.
- `planned`: `progress.json` has a builder-ready row.
- `vague`: represented only by umbrella prose; planner must split it.
- `missing`: upstream behavior has no Go target or row.
- `owned`: intentional Go-native divergence with documented contract.
- `excluded`: upstream surface is not runtime/product behavior.

## Output Files

For planning work, update these together:

- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`;
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`;
- `docs/content/building-goscrapling/architecture_plan/progress.json`.

## Validation

Run `go run ./cmd/progress validate`, `go run ./cmd/progress write`, and
`git diff --check` after ledger or progress edits. Run `go test ./... -count=1`
when tests or generated progress docs can observe the change.

## Done

A parity pass is done when every mapped behavior has a status, evidence path,
progress anchor or owned/excluded decision, and explicit next skill route.
