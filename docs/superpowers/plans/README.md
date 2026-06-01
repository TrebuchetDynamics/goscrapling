# Superpowers Plans

Implementation plans are grouped by the responsibility they change so builders can find the plan that matches the code boundary they are about to touch.

## Boundaries

- `parser-foundation/` owns first-principles parser and adaptive-storage implementation plans. It exposes parser-focused builder steps and must never grow CLI, browser, spider, or control-plane orchestration details.
- `control-plane/` owns progress-ledger, queue generation, and upstream-map planning. It exposes operator/control-plane builder steps and must never depend on runtime implementation internals beyond documented validation commands.
- `validation/` owns cross-layer and end-to-end proof plans. It exposes fixture-backed validation harness work and must never require live web access, credentials, or unplanned production features.
- `architecture-deepening/` owns no-behavior-change refactor plans for existing runtime seams, grouped by runtime boundary (`browser-evidence/`, `spider-runtime/`, and `static-extract-cli/`). It exposes characterization-test-first deepening steps and must never change public APIs or product behavior silently.
