# Skill Builder Handoff

This page is generated from `meta.builder_loop` in the canonical progress
file. Keep shared skill handoff facts in `progress.json`; keep row-specific
execution facts on the rows.

<!-- PROGRESS:START kind=builder-loop-handoff -->
## Control Plane

- Entrypoint: `.pi/skills/goscrapling-skill-manager/SKILL.md`
- Plan: `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- Coverage ledger: `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`
- Agent queue: `docs/content/building-goscrapling/builder-loop/agent-queue.md`
- Progress schema: `docs/content/building-goscrapling/builder-loop/progress-schema.md`
- Candidate source: `docs/content/building-goscrapling/architecture_plan/progress.json`
- Unit tests: `go test ./... -count=1`

## Candidate Policy

- Prefer P0 planned rows with source_refs, write_scope, test_commands, acceptance, and done_signal.
- Do not assign umbrella rows until a planner splits them.
- Every implementation row starts with a failing test or fixture update.
- Update progress.json and parity docs in the same slice that changes behavior.
<!-- PROGRESS:END -->
