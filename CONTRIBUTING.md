# Contributing

`goscrapling` is a Go-native Scrapling-style feature port. Feature work should
stay tied to the port plan instead of becoming a parallel scraper backlog.

## Planning Sources

Use these files before splitting or building work:

- [Scrapling Feature Map](docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md)
- [Upstream Coverage Ledger](docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md)
- [Progress Ledger](docs/content/building-goscrapling/architecture_plan/progress.json)
- [Agent Queue](docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md)
- [Progress Schema](docs/content/building-goscrapling/builder-loop/schema/progress-schema.md)

Before adding behavior, update or pick a builder-sized row in `progress.json`
with source refs, write scope, tests, acceptance, and a done signal. Do not
create a parallel backlog.

Regenerate queue surfaces after progress-ledger edits:

```sh
go run ./cmd/progress write
```

## Validation

Run the full hermetic validation suite before claiming a slice is complete:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Run the deterministic full local CLI smoke suite:

```sh
go test ./cmd/goscrapling -run TestGoscraplingFullLocalEndToEnd -count=1
```

Run the optional live practice-site suite only when live network access and
robots.txt preflight are acceptable:

```sh
GOSCRAPLING_LIVE_E2E=1 go test ./cmd/goscrapling -run TestLivePracticeSitesEndToEnd -count=1 -timeout 10m
```

## Reference Checkout

The upstream Scrapling repository may be cloned locally at:

```text
references/Scrapling
```

Keep that checkout ignored by git. It is study material, not vendored source.
