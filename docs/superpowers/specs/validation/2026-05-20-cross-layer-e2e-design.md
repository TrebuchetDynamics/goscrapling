# Cross-Layer Local E2E Design

## Goal

Increase goscrapling end-to-end confidence across the implemented layers without expanding runtime behavior beyond existing Scrapling-style parity rows.

## Scope

Add one builder-ready progress row and one hermetic cross-layer E2E test that connects currently implemented layers:

- local HTTP fixture via `httptest.Server`;
- static fetcher request options and response metadata;
- parser CSS extraction and adaptive selector save/relocate;
- persistent adaptive storage through a temp directory;
- spider crawl/follow/session flow;
- CLI binary extraction output;
- Gormes browser extraction adapter through a fake browser engine.

Out of scope: live web access, credentials, real browser smoke tests, proxy rotation, stealth behavior, robots crawling policy changes, and new production features.

## Architecture

The E2E test lives in `cmd/goscrapling/cross_layer_e2e_test.go` because the existing binary-build helpers are already in that package. Package-level fixture helpers keep the test hermetic and readable. The test does not add new production APIs; it proves that public package boundaries compose correctly from local fixtures.

## Progress Tracking

`docs/content/building-goscrapling/architecture_plan/progress.json` remains the source of truth. Add a planned row named `Cross-layer local E2E validation harness` under `phase-5-cli-tooling/tool-surfaces`, with write scope limited to the progress ledger, generated builder-loop docs, and the new E2E test.

## Testing

Focused test command:

```sh
go test ./cmd/goscrapling -run TestGoscraplingCrossLayerLocalEndToEnd -count=1
```

Completion gate:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress write
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```
