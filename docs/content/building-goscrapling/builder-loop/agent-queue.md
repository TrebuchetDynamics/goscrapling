# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Selector generation helpers

- Phase: `phase-0-parser-adaptive / static-document`
- Priority: `P2`
- Owner: `parser`
- Size: `medium`
- Contract status: `draft`
- Contract: Generate robust CSS, full CSS, XPath, and full XPath selectors for selected elements using a Go-native equivalent of Scrapling selector generation.
- Ready when: Traversal helpers expose enough DOM context to generate selectors.
- Write scope: `selector_generation.go`, `selector_generation_test.go`
- Test commands: `go test ./... -run TestSelectorGeneration -count=1`
- Acceptance: Fixtures prove generated selectors re-select the original element and remain deterministic.
- Done signal: Selector generation tests pass.
- Source refs: `references/Scrapling/scrapling/core/mixins.py`, `references/Scrapling/docs/api-reference/selector.md`
<!-- PROGRESS:END -->
