# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Browser sessions and page pool lifecycle

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `draft`
- Contract: Add DynamicSession-style lifecycle with reusable browser contexts/pages, page pool stats, and clean close semantics.
- Ready when: Real browser adapter or fake-backed engine seam is stable.
- Write scope: `engines/browser/browser_session.go`, `engines/browser/browser_session_test.go`, `docs/content/building-goscrapling/architecture_plan/progress.json`, `docs/content/building-goscrapling/builder-loop/agent-queue.md`, `docs/content/building-goscrapling/builder-loop/blocked-slices.md`, `docs/content/building-goscrapling/builder-loop/next-slices.md`
- Test commands: `go test ./... -run TestBrowserSessionPool -count=1`
- Acceptance: Fake and local fixtures prove session reuse, pool stats, busy/free/error page counts, and close behavior.
- Done signal: Browser session pool tests pass.
- Source refs: `references/Scrapling/scrapling/fetchers/chrome.py`, `references/Scrapling/scrapling/engines/_browsers/_base.py`, `references/Scrapling/scrapling/engines/_browsers/_controllers.py`, `engines/browser/browser.go`, `engines/browser/browser_adapter.go`
<!-- PROGRESS:END -->
