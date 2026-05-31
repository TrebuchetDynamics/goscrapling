# goscrapling Project Skill Inventory

These repo-local workflow skills live under `.pi/skills/` so Pi discovers them
as project skills at startup. They are source-controlled project instructions,
not global user skills and not a parallel backlog.

## Required Order

1. `goscrapling-skill-manager` — route the request and scope the pass.
2. `goscrapling-scrapling-parity` — map upstream behavior when parity is unclear.
3. `goscrapling-planner` — create or refine one builder-ready row.
4. `goscrapling-tdd-slice` — write the failing test for runtime behavior.
5. `goscrapling-builder` — implement exactly one row and validate it.

Use `scrapling-mirror` only as an orchestrator for mirror refresh, next-row
selection, and repeated safe slices. It must call into the narrower skills when
parity mapping, planning, TDD, or implementation work starts.

Docs-only skill maintenance may stay in `goscrapling-skill-manager`, but must
still preserve `progress.json` as the only implementation backlog.

## Skill File Standard

Each `SKILL.md` should include YAML frontmatter with `name` and `description`,
a short purpose, inputs, rules, validation, and done criteria. Descriptions
should state when to use the skill, not summarize the whole workflow.

## Validation

For skill-doc-only changes, run:

```sh
git diff --check
```

For progress, feature-map, coverage-ledger, generated-doc, schema, or runtime
changes, also run the relevant commands from `AGENTS.md` and the selected skill.
