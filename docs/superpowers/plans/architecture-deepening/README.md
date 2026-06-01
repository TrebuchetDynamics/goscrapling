# Architecture Deepening Plans

No-behavior-change refactor plans are grouped by the runtime boundary they deepen. Each subfolder owns characterization-first implementation plans for one code responsibility and must not silently change public APIs, product behavior, data formats, or validation gates.

## Boundaries

- `browser-evidence/` owns rendered browser/Gormes HTML evidence extraction deepening. It exposes plans for shared parsing and evidence value construction; it must never own spider crawl scheduling, CLI command routing, or progress-ledger control-plane behavior.
- `spider-runtime/` owns spider crawl engine orchestration deepening. It exposes plans for scheduler/runtime state, task accounting, and crawl result mutation; it must never own browser evidence parsing, CLI command routing, or fetcher/browser adapters outside documented spider seams.
- `static-extract-cli/` owns static extract CLI command deepening. It exposes plans for parse/plan/execute/render/write command seams; it must never own parser internals, browser runtime behavior, or spider scheduling.
