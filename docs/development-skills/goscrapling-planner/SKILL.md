# goscrapling-planner

Use this to turn a parity gap into one builder-ready row.

## Inputs

- User goal.
- `progress.json`.
- Scrapling source refs.
- Existing goscrapling tests and APIs.

## Row Contract

A builder-ready row needs:

- `name`
- `status`
- `priority`
- `contract`
- `contract_status`
- `slice_size`
- `execution_owner`
- `source_refs`
- `ready_when`
- `not_ready_when` when broad or risky
- `write_scope`
- `test_commands` or `no_test_required`
- `acceptance`
- `done_signal`

## Planning Rules

- Prefer one vertical slice that proves behavior end to end.
- Split umbrella rows before handing work to a builder.
- Preserve Go-native API design, but cite the Scrapling behavior being kept.
- Keep tests hermetic: `httptest`, temp dirs, fake browser engines, and local
  fixtures before live network or browser tests.
- Update the feature map and coverage ledger when a new upstream source class
  becomes relevant.

## Done

The row is ready when another agent can implement it without rediscovering the
upstream behavior or guessing the write scope.
