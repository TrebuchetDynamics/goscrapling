# goscrapling Control Plane Design

Date: 2026-05-13

Status: approved by Juan as approach A.

## Decision

`goscrapling` will adopt a small Gormes-style porting control plane before the
next runtime feature slice.

The goal is to make the Scrapling port durable enough for long-term work:
future agents should not guess what to build, claim parity from prose, or turn
the repository into a generic Go scraping helper. The repo must keep one
canonical backlog, generated execution views, upstream drift checks, and
explicit behavior atoms for the Python-to-Go mapping.

## Source Lesson From Gormes

Gormes made the Hermes Python-to-Go port tractable by adding process and test
surfaces around the port:

- `progress.json` is the only backlog.
- A progress validator rejects vague active work.
- Generated docs expose the next assignable rows, blocked rows, and umbrella
  rows that need splitting.
- Upstream coverage tests fail when source classes are not mapped.
- Parity manifests freeze public surfaces before handler implementation.
- Boundary docs keep donor compatibility namespaces from becoming permanent
  product architecture.

goscrapling should copy this control-plane pattern, scaled down for a young
library.

## Scope

This design covers planning infrastructure only. It does not implement
`Response`, `Fetcher`, browser fetching, spiders, or CLI runtime behavior.

The first implementation plan should produce:

- a small `internal/progress` package;
- a `cmd/progress` validator/generator;
- generated builder-loop docs;
- a stronger Scrapling upstream coverage test;
- a behavior-atoms document for the first porting surfaces;
- a boundary document for Scrapling compatibility versus Go-native packages;
- progress row refinements that make Phase 1 builder-ready.

## Non-Goals

- No live web scraping.
- No real browser dependency.
- No full Gormes progress engine copy.
- No site generator or landing-page integration.
- No submodule or parent fleet restructuring beyond recording the submodule
  pointer after the goscrapling commit.

## Architecture

### Progress Model

Create `internal/progress` as a small package that understands the existing
`docs/content/building-goscrapling/architecture_plan/progress.json` shape.

It should own:

- JSON loading;
- row validation;
- derived status counts;
- selection of builder-ready rows;
- selection of blocked rows;
- selection of umbrella rows;
- markdown rendering for generated docs.

The model should stay intentionally smaller than Gormes. It does not need
health metadata, planner verdicts, site sync, or subagent execution history.

### Progress CLI

Create `cmd/progress` with two commands:

```sh
go run ./cmd/progress validate
go run ./cmd/progress write
```

`validate` parses `progress.json`, checks schema invariants, and prints a short
summary.

`write` validates first, then regenerates markered sections in builder-loop
docs. It should fail if a target file is missing a marker, because silent
partial generation would make the queue untrustworthy.

### Generated Builder-Loop Docs

Add generated pages under:

```text
docs/content/building-goscrapling/builder-loop/
```

Required pages:

- `builder-loop-handoff.md`
- `agent-queue.md`
- `next-slices.md`
- `blocked-slices.md`
- `umbrella-cleanup.md`

The generated views are read-only operator surfaces. `progress.json` remains
the source of truth.

### Behavior Atoms

Add:

```text
docs/content/building-goscrapling/architecture_plan/scrapling-behavior-atoms.md
```

This file records the smallest meaningful Scrapling behaviors that need Go
parity. The first version should cover:

- `Response` metadata plus selector behavior;
- response body/json helpers;
- static `Fetcher` methods;
- `FetcherSession` option merging, cookies, redirects, timeout, and retries;
- browser fetcher contract shape;
- spider request, scheduler, session routing, response follow, and crawl
  result behavior.

Each atom should include upstream refs, visible contract, Go target,
progress row, validation command, and status.

### Boundary Document

Add:

```text
docs/content/building-goscrapling/architecture_plan/boundaries.md
```

The key rule:

- `compat/scrapling` is acceptable for evidence, drift manifests, and explicit
  compatibility shims.
- Durable Go architecture belongs in Go-native package names such as
  `adaptive`, `fetcher`, `browser`, `spider`, `storage`, and `cmd/goscrapling`.

This mirrors Gormes' `internal/hermes` boundary: donor language is useful while
proving parity, but it should not become the permanent architecture.

### Coverage Checks

Strengthen `progress_docs_test.go` so it checks more than a fixed list of
strings. If `references/Scrapling` is present, the test should inventory core
upstream source classes and fail when a class is neither represented nor
explicitly excluded in the coverage ledger.

It does not need full nested symbol coverage yet. The immediate target is to
catch new top-level Scrapling source areas that would otherwise be ignored.

### Phase 1 Row Split

Refine Phase 1 from two broad draft rows into smaller rows:

- `Response metadata and selector contract`
- `Response body, text, bytes, and JSON helpers`
- `Static Fetcher method surface over net/http`
- `FetcherSession option merging and cookies`
- `Redirect, timeout, retry, and error taxonomy`

The first row should be marked `fixture_ready` only when its test command and
write scope are precise enough for a TDD implementation pass.

## Data Flow

1. Planner updates `progress.json`.
2. `go run ./cmd/progress validate` checks the ledger.
3. `go run ./cmd/progress write` regenerates builder-loop docs.
4. Builder reads `agent-queue.md` or a user-named row.
5. Builder writes a failing test from the row contract.
6. Builder implements the row, updates `progress.json`, and regenerates docs.

No other backlog is allowed.

## Error Handling

Validation should report all row-level problems it can find in one run:

- invalid status;
- invalid contract status;
- missing contract;
- missing source refs;
- missing write scope;
- missing tests or no-test reason;
- missing acceptance;
- missing done signal;
- complete row without `contract_status: validated`;
- `in_progress` row with `slice_size: umbrella`.

Generation should fail when:

- progress validation fails;
- a target markdown file cannot be read;
- a target markdown file lacks the requested marker;
- writing a generated file fails.

## Testing

Required validation after the slice:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress write
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Focused tests should include:

- progress validation rejects incomplete contract rows;
- progress validation accepts the current ledger;
- generated queue excludes complete rows and umbrella rows;
- generated umbrella cleanup includes the crawler production-controls umbrella;
- upstream coverage test fails on unmapped Scrapling source classes when the
  local checkout is present.

## Acceptance Criteria

This control-plane upgrade is complete when:

- `cmd/progress validate` succeeds on the repo's canonical `progress.json`;
- `cmd/progress write` regenerates all builder-loop docs deterministically;
- `go test ./... -count=1` passes;
- README and AGENTS mention the progress CLI;
- Phase 1 has small builder-ready rows instead of only broad draft rows;
- the next runtime implementation can start from `agent-queue.md` without
  rediscovering the Scrapling behavior.

## Follow-Up

After this control plane is in place, the next runtime work should be the first
Phase 1 row: `Response metadata and selector contract`, implemented through
TDD against local fixtures.
