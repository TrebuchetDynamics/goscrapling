# goscrapling-scrapling-parity

Use this when mapping Scrapling behavior into goscrapling.

## Goal

Extract behavior atoms from upstream Scrapling and classify the Go status as
`covered`, `partial`, `planned`, `vague`, `missing`, `owned`, or `excluded`.

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

1. `references/Scrapling/scrapling/**`
2. `references/Scrapling/docs/**`
3. Existing goscrapling tests and docs
4. `progress.json`

## Output

For planning work, update these files together:

- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- `docs/content/building-goscrapling/architecture_plan/progress.json`

Do not claim parity from prose alone. Covered behavior needs tests.
