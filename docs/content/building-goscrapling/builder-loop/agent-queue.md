# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Stealth browser controls and fingerprint options

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P2`
- Owner: `browser`
- Size: `medium`
- Contract status: `draft`
- Contract: Expose stealth controls separately from normal browser fetching: fingerprint/header generation, WebRTC blocking, canvas hiding, WebGL control, extra args, real Chrome, and proxy hygiene.
- Ready when: Normal browser context options are stable.
- Write scope: `browser_stealth.go`, `browser_stealth_test.go`, `docs/content/building-goscrapling/architecture_plan/boundaries.md`
- Test commands: `go test ./... -run TestStealthBrowserControls -count=1`
- Acceptance: Fixtures prove controls are explicit, unsupported anti-bot claims are rejected visibly, and docs describe supported behavior.
- Done signal: Stealth browser control tests pass.
- Source refs: `references/Scrapling/scrapling/fetchers/stealth_chrome.py`, `references/Scrapling/scrapling/engines/_browsers/_stealth.py`, `references/Scrapling/scrapling/engines/toolbelt/fingerprints.py`, `references/Scrapling/docs/fetching/stealthy.md`
<!-- PROGRESS:END -->
