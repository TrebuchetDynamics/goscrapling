---
name: scrapling-mirror
description: "Use when running goscrapling Scrapling-parity automation: refresh the upstream mirror matrix, select the next progress.json row, write the failing test, build one validated slice, and optionally continue safe slices."
---

# scrapling-mirror

## Purpose

Run the goscrapling mirror-and-build loop without creating a parallel backlog.
This skill orchestrates the repo-local parity, planner, TDD, and builder skills;
it does not bypass `progress.json`, does not claim full parity, and does not
turn unsafe browser/stealth/proxy behavior into incidental work.

## Modes

Parse the user request into one mode:

- `mirror-only` — refresh the Scrapling -> goscrapling matrix and upstream app
  map without runtime implementation.
- `build-next` — refresh mirror evidence, choose one assignable row, write the
  failing test, implement it, and validate.
- `loop` — repeat `build-next` while validation is green and no stop condition
  fires. Use only when the user asks for automatic continuation.
- `status` — report upstream delta, mirror status, next assignable rows, and
  validation receipts without editing.

Default to `build-next` when the user asks to "execute", "build", "autopilot",
"continue parity", or "make goscrapling automatically".

## Required Inputs

Read or run these before editing:

- `AGENTS.md`;
- `.pi/skills/goscrapling-skill-manager/references/project-skill-inventory.md`;
- `docs/content/building-goscrapling/architecture_plan/progress.json`;
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`;
- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`;
- `docs/content/building-goscrapling/architecture_plan/upstream-app-map.json`;
- `docs/research/scrapling-parity-matrix.md`;
- `git status --short --branch --untracked-files=all`.

If `codemap.md` is absent, say so and continue from the ledgers plus nearest
package docs/tests. If `references/Scrapling` is absent, skip upstream fetches,
report the missing checkout, and do not invent source refs.

## Preflight

For deterministic preflight, run:

```sh
.pi/skills/scrapling-mirror/scripts/preflight.sh
```

The script prints repo status, fetches the upstream Scrapling remote when
available, shows upstream deltas, validates app-map/progress JSON, and shows the
current generated queue surfaces. Treat script output as evidence, not as
permission to edit unrelated dirty files.

## Workflow

1. Route through `goscrapling-skill-manager` and state the selected mode.
2. Inspect dirty worktree changes. Continue only when they are in-scope or
   clearly generated/local state; otherwise ask one ownership question.
3. Mirror upstream:
   - run `git -C references/Scrapling fetch --tags origin` when the checkout is
     present and network is allowed;
   - compare `HEAD..origin/main` and classify changed files as source, tests,
     docs, assets, or packaging;
   - update the mirror matrix, upstream app map, feature map, coverage ledger,
     and progress rows together when the delta is feature-bearing;
   - run `go run ./cmd/progress map-write` and `go run ./cmd/progress write`
     when generated surfaces change.
4. Pick the next slice from canonical evidence:
   - prefer `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md` and
     `agent-queue.md`;
   - only take rows that are planned or in_progress, non-umbrella, unblocked,
     and have source refs, write scope, test commands, acceptance, and done
     signal;
   - if no row is builder-ready, load `goscrapling-planner` and make exactly one
     row ready.
5. For runtime behavior, load `goscrapling-tdd-slice`, write the smallest
   failing focused test, and record the expected failure.
6. Load `goscrapling-builder` and implement exactly that row within write scope.
7. Validate, then update `progress.json` and generated docs only after evidence
   is green.
8. In `loop` mode, choose the next bounded row and continue only if validation
   is green, ownership is clear, and context/budget allows.

## Validation Gate

Before reporting a slice complete, run the commands named on the row plus:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

When mirror/app-map files changed, also run:

```sh
go run ./cmd/progress map-validate
jq empty docs/content/building-goscrapling/architecture_plan/upstream-app-map.json
```

When generated queue surfaces changed, run:

```sh
go run ./cmd/progress write
```

## Stop Conditions

Stop and report `blocked` instead of continuing when:

- dirty files appear unrelated or user-owned;
- the next row is umbrella-sized or lacks builder-ready metadata;
- a failing test forces work outside the row write scope;
- live web, live browser, live LLM, credentials, deploys, publishing, or money
  would be required for the core proof;
- stealth, anti-bot bypass, proxy rotation, or browser automation behavior is
  not explicitly named by the row, tests, and docs;
- validation fails twice with the same blocker;
- context or soft token budget is too low for another safe slice.

## Output Contract

End each run with:

- mode and selected row;
- upstream snapshot/delta;
- files changed;
- validation commands and outcomes;
- next row candidate or blocker;
- whether continuation happened or stopped.
