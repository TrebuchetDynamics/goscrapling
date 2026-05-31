---
name: goscrapling-skill-manager
description: Use when starting goscrapling work, choosing a repo-local workflow, editing goscrapling skills, or deciding whether a request needs parity, planning, TDD, builder, or docs-only routing.
---

# goscrapling-skill-manager

## Purpose

Route every goscrapling task through the smallest workflow that keeps the repo
moving toward tested Scrapling-style parity. Do not let goscrapling drift into a
generic scraper helper library or a side-backlog-driven project.

## Required Inputs

Read these before substantive planning or implementation:

- user request;
- `AGENTS.md`;
- `references/project-skill-inventory.md` (relative to this skill directory);
- `docs/content/building-goscrapling/architecture_plan/progress.json`;
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`;
- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`.

If `codemap.md` is absent, say so in the report and continue from the files
above plus the nearest package docs/tests.

## Routing

| Request shape | Required next skill |
|---|---|
| Unsure, broad, or mixed request | Stay here, split scope, then route |
| "What does Scrapling do?" or parity inventory | `goscrapling-scrapling-parity` |
| "What should we build next?" or row readiness | `goscrapling-planner` |
| Mirror matrix, autopilot, "build goscrapling automatically", or repeated safe slices | `scrapling-mirror`, then the narrower parity/planner/TDD/builder skills as needed |
| Implement a runtime behavior row | `goscrapling-tdd-slice`, then `goscrapling-builder` |
| Edit repo-local skills | This skill plus skill-authoring checks; keep changes docs-only unless asked |

## Rules

- No side backlog: new work must reference or update `progress.json`.
- No broad builder work: split umbrella rows before implementation.
- No live web tests for core behavior: prefer `httptest`, temp dirs, fake
  browser engines, and fixtures.
- Keep `references/Scrapling` ignored study material.
- Skill edits must improve future routing or reduce ambiguity; do not add
  narrative session history.

## Done

A routing pass is done when the report names the selected skill path, exact
scope, files inspected, and whether runtime validation was needed or skipped.
