# Progress Schema

`progress.json` is the canonical work ledger for the port. It is intentionally
modeled after the Gormes port ledger, but smaller.

## Item Fields

| Field | Required when | Meaning |
|---|---|---|
| `name` | every item | Human-readable row name. |
| `status` | every item | `planned`, `in_progress`, or `complete`. |
| `priority` | optional | `P0` through `P4`; `P0` rows are product-critical. |
| `contract` | every implementation row | The upstream Scrapling behavior or owned Go-native behavior. |
| `contract_status` | contract rows | `missing`, `draft`, `fixture_ready`, or `validated`. |
| `slice_size` | contract rows | `small`, `medium`, `large`, or `umbrella`. |
| `execution_owner` | contract rows | Logical area such as `parser`, `storage`, `fetcher`, `browser`, `spider`, `cli`, `integration`, or `docs`. |
| `source_refs` | contract rows | Upstream and local files used to derive the row. |
| `ready_when` | contract rows | Conditions that make the row safe to assign. |
| `not_ready_when` | broad or risky rows | Conditions that make the row too broad or unsafe. |
| `write_scope` | contract rows | Files or package areas a builder may edit. |
| `test_commands` | contract rows | Commands that prove the row. Use `no_test_required` only for rare docs-only rows. |
| `acceptance` | contract rows | Testable done criteria. |
| `done_signal` | contract rows | Evidence that the row can close. |

## Rules

- `progress.json` is not a side note; it is the source of truth.
- `in_progress` rows cannot be `slice_size: umbrella`.
- A builder should not take a row without source refs, write scope, tests,
  acceptance, and done signal.
- Complete contract rows must use `contract_status: validated`.
- Broad rows stay planned until split into builder-sized rows.
- New upstream feature-bearing files require updates to the coverage ledger,
  feature map, and progress ledger in the same planning pass.

## Validation

Run:

```sh
go test ./... -count=1
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
```

The Go test checks that the progress ledger has valid statuses, contract rows
have handoff metadata, complete rows are validated, and the upstream coverage
ledger names the core Scrapling source classes when the local upstream checkout
is present.
