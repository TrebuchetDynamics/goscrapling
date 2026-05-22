# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Benchmarks and parity scorecard

- Phase: `phase-5-cli-tooling / tool-surfaces`
- Priority: `P4`
- Owner: `docs`
- Size: `medium`
- Contract status: `draft`
- Contract: Add benchmark coverage and a generated parity scorecard so claims such as coverage, speed, and feature completion remain evidence-backed.
- Ready when: Major subsystem APIs are stable enough to benchmark.
- Write scope: `benchmarks/`, `docs/research/`, `cmd/progress/`
- Test commands: `go test ./... -run TestParityScorecard -count=1`
- Acceptance: Benchmark fixtures and generated scorecard summarize parser/fetcher/spider/CLI coverage without live services.
- Done signal: Parity scorecard tests pass.
- Source refs: `references/Scrapling/docs/benchmarks.md`, `references/Scrapling/pyproject.toml`, `docs/research/scrapling-parity-matrix.md`
<!-- PROGRESS:END -->
