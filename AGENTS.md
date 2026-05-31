# AGENTS.md - goscrapling

This repository is a Go-native Scrapling-style feature port.

## Port Rule

`goscrapling` is only worth building if it keeps moving toward real Scrapling
feature parity. Do not describe it as a complete port until parser, adaptive
storage, fetchers, browser fetchers, spiders, CLI surfaces, and tool
integration are backed by tests and progress rows.

Use D4Vinci/Scrapling as the parity oracle:

- Local study checkout: `references/Scrapling`
- Observed upstream commit: `6380ef0f266a5fff898c18953d6b03ca320b2fd4`
- Observed upstream release: `v0.4.8`

This project is not affiliated with D4Vinci/Scrapling.

## Single Source Of Truth

Use these files before splitting or building work:

- `docs/content/building-goscrapling/architecture_plan/progress.json`
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`
- `docs/content/building-goscrapling/builder-loop/progress-schema.md`

Do not create parallel backlogs. New feature work must either update an
existing `progress.json` row or add a builder-sized row with source refs, write
scope, tests, acceptance, and a done signal.

## Repo-Local Workflow Skills

Follow these repo-local Pi project skills as operating docs. They live under
`.pi/skills/` so Pi discovers them at startup:

- `.pi/skills/goscrapling-skill-manager/SKILL.md`
- `.pi/skills/goscrapling-scrapling-parity/SKILL.md`
- `.pi/skills/goscrapling-planner/SKILL.md`
- `.pi/skills/goscrapling-builder/SKILL.md`
- `.pi/skills/goscrapling-tdd-slice/SKILL.md`
- `.pi/skills/scrapling-mirror/SKILL.md` (orchestrates mirror refresh and safe repeated slices; it must still use the narrower skills above for parity, planning, TDD, and builder work)

The expected flow is:

1. Route the work through the skill manager.
2. Use the parity skill to identify upstream behavior atoms.
3. Use the planner skill to create or refine one builder-ready row.
4. Use the TDD slice skill to write a failing test first.
5. Use the builder skill to implement only that row.

## Validation

Run these checks before claiming a slice is complete:

```sh
go test ./... -count=1
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

For a docs-only progress change, `go test ./... -count=1` still matters because
`progress_docs_test.go` validates the progress ledger and upstream coverage
ledger.

After editing `progress.json`, regenerate the queue surfaces:

```sh
go run ./cmd/progress write
```

## Boundaries

- Keep `references/Scrapling` ignored by git. It is study material, not
  vendored source.
- Prefer Go-native APIs over copying Python names blindly, but preserve
  Scrapling-visible behavior unless a progress row marks a narrow owned
  divergence.
- Use hermetic tests and local HTTP/browser fixtures before any live web
  dependency.
- Do not add stealth, anti-bot bypass, proxy rotation, or browser automation
  behavior without explicit tests, operator-visible controls, and docs.
