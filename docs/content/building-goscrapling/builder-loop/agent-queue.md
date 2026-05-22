# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Browser context options and resource blocking

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `draft`
- Contract: Support browser context options including cookies, locale, timezone, user agent, extra headers, proxy, CDP/real Chrome selection, resource disabling, custom blocked domains, ad-domain blocking, and DNS-over-HTTPS flag.
- Ready when: Browser adapter and session rows are stable.
- Write scope: `browser_options.go`, `browser_options_test.go`, `browser_resources_test.go`
- Test commands: `go test ./... -run TestBrowserContextOptions -count=1`
- Acceptance: Fake engine fixtures prove every option is explicit, validated, and passed through without hidden stealth behavior.
- Done signal: Browser context option tests pass.
- Source refs: `references/Scrapling/scrapling/engines/_browsers/_config_tools.py`, `references/Scrapling/scrapling/engines/toolbelt/navigation.py`, `references/Scrapling/scrapling/engines/toolbelt/ad_domains.py`, `references/Scrapling/docs/fetching/dynamic.md`

## 2. Browser wait conditions, page actions, downloads, screenshots, and XHR capture

- Phase: `phase-3-browser / browser-fetcher`
- Priority: `P1`
- Owner: `browser`
- Size: `medium`
- Contract status: `draft`
- Contract: Port dynamic fetch controls for wait selector/state, network idle, fixed wait, page actions, downloads, screenshots, and captured XHR responses.
- Ready when: Browser session pool is stable.
- Write scope: `browser.go`, `browser_actions.go`, `browser_actions_test.go`, `testdata/browser/`
- Test commands: `go test ./... -run TestBrowserActionsAndCapture -count=1`
- Acceptance: Fixtures prove waits/actions/capture fields are passed to the engine and reflected in Response metadata.
- Done signal: Browser action and capture tests pass.
- Source refs: `references/Scrapling/scrapling/engines/_browsers/_page.py`, `references/Scrapling/scrapling/engines/toolbelt/convertor.py`, `references/Scrapling/docs/fetching/dynamic.md`
<!-- PROGRESS:END -->
