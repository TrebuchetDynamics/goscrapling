# goscrapling-skill-manager

Use this before planning or building goscrapling work.

## Purpose

Route work so the project keeps moving toward a real Scrapling-style Go port
instead of becoming an unrelated scraper helper library.

## Required Inputs

- User request.
- `AGENTS.md`.
- `docs/content/building-goscrapling/architecture_plan/progress.json`.
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`.
- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`.

## Routing

- If the request asks what Scrapling does, use `goscrapling-scrapling-parity`.
- If the request asks what to build next, use `goscrapling-planner`.
- If the request asks to implement a row, use `goscrapling-tdd-slice` and then
  `goscrapling-builder`.
- If the request asks for broad architecture, update the feature map and
  progress rows before runtime code.

## Rules

- No side backlog. Work must update or reference `progress.json`.
- No broad builder work. Split umbrella rows before implementation.
- No live web tests for core behavior. Use `httptest`, temp dirs, fake browser
  engines, and fixture servers first.
- Keep `references/Scrapling` as ignored study material.
