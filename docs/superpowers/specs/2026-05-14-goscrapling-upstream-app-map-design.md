# goscrapling Upstream App Map Design

Date: 2026-05-14

Status: approved by Juan.

## Decision

`goscrapling` will add a full upstream Scrapling application map before the
next runtime feature slice.

The map will cover upstream Scrapling source files, tests, and behavior-bearing
docs. Its purpose is to make the Python-to-Go port systematic: every upstream
surface should have an explicit Go target, parity status, progress row, and
test evidence path before builders claim parity.

Generated Python-to-Go translations are allowed only as static reference
evidence. They are not production code and must not be copied into runtime Go
packages without going through the normal progress-row and TDD workflow.

## Scope

The first slice will produce:

- `docs/content/building-goscrapling/architecture_plan/upstream-app-map.json`
- `docs/content/building-goscrapling/architecture_plan/upstream-app-map.md`
- py2many static reference outputs under `docs/research/python-to-go-probes/`
- app-map validation and rendering code in `internal/progress`
- `cmd/progress map-validate`
- `cmd/progress map-write`

The JSON manifest is the source of truth. The Markdown file is generated for
human review.

## Non-Goals

- Do not add py2many, pytago, or other translators as Go module dependencies.
- Do not commit generated Go into production packages.
- Do not claim parity from py2many output.
- Do not split or implement runtime feature rows in this slice.
- Do not run live web or browser tests.

## Manifest Shape

Each app-map entry will link an upstream artifact to the local port plan:

- upstream ref
- kind: `source`, `test`, `doc`, or `asset`
- upstream symbols when cheaply discoverable
- feature-map anchor
- behavior atom names
- Go target package or file area
- progress row name
- coverage status: `covered`, `partial`, `planned`, `vague`, `owned`, or
  `excluded`
- translation suitability: `probe_candidate`, `manual_rewrite`, `not_useful`,
  or `not_applicable`
- static reference paths for py2many output when available
- notes for gaps, owned divergence, or exclusion

The first map should be explicit enough to validate every feature-bearing
upstream source, test, and doc file. It can group non-behavior assets and
translated docs as excluded entries.

## py2many Static Reference Pass

Run py2many before finalizing the first manifest so the map can record whether
translation output is useful for each upstream area.

The py2many pass should:

- target a representative set first, then expand if the command is cheap;
- write output only under `docs/research/python-to-go-probes/py2many/`;
- include a short generated summary with command, version, inputs, failures,
  and notable output paths;
- mark generated output as reference-only in file headers or adjacent notes;
- tolerate missing py2many by recording that the probe could not run.

The useful result is not idiomatic Go. The useful result is a static aid for
finding structs, methods, obvious control flow, unsupported dynamic Python, and
manual rewrite hotspots.

## Progress Control Plane

Extend `cmd/progress` instead of adding a separate tool:

```sh
go run ./cmd/progress map-validate
go run ./cmd/progress map-write
```

`map-validate` should load `upstream-app-map.json`, verify required fields, and
check that every upstream source/test/doc entry that should be mapped is present
when `references/Scrapling` exists.

`map-write` should validate first, render `upstream-app-map.md`, and print the
files it regenerated.

The existing `validate` and `write` commands keep their current behavior. The
app map becomes an additional parity gate, not a replacement for
`progress.json`.

## Data Flow

1. Inventory `references/Scrapling/scrapling`, `references/Scrapling/tests`,
   and behavior-bearing `references/Scrapling/docs`.
2. Run py2many against selected source entries and store reference output under
   `docs/research/python-to-go-probes/py2many/`.
3. Create or update `upstream-app-map.json`.
4. Run `go run ./cmd/progress map-validate`.
5. Run `go run ./cmd/progress map-write`.
6. Future planner work uses the map to split vague rows into builder-sized
   progress rows.

## Testing

Focused tests should cover:

- app-map JSON loading;
- required field validation;
- source/test/doc inventory coverage;
- generated Markdown rendering;
- `cmd/progress map-validate`;
- `cmd/progress map-write`.

Full validation before closing the slice:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress map-validate
go run ./cmd/progress write
go run ./cmd/progress map-write
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
jq empty docs/content/building-goscrapling/architecture_plan/upstream-app-map.json
git diff --check
```

## Acceptance Criteria

- The full upstream Scrapling app is represented by a machine-readable map.
- Every feature-bearing source/test/doc entry has a local target, progress
  anchor, status, and evidence path or an explicit exclusion.
- py2many output is available as reference evidence where it successfully ran.
- Generated Markdown gives humans a readable view of the same map.
- The progress control plane can validate and regenerate the map.
- No production Go code is generated by or copied from py2many in this slice.
