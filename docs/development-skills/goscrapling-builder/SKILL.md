# goscrapling-builder

Use this to implement exactly one builder-ready row.

## Inputs

- One `progress.json` row.
- Failing test from `goscrapling-tdd-slice`.
- Source refs listed on the row.

## Implementation Rules

- Touch only the row write scope unless the failing test proves the scope is
  wrong.
- Prefer standard-library Go and existing repository patterns.
- Keep APIs small and stable before adding convenience layers.
- Do not add live network dependencies to core tests.
- Do not implement browser stealth, proxy rotation, or crawling controls as
  incidental side effects of another row.

## Completion

Before closing the row, run:

```sh
go test ./... -count=1
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Then update `progress.json`:

- set `status` to `complete`;
- set `contract_status` to `validated`;
- ensure acceptance and done signal describe the evidence that passed.
