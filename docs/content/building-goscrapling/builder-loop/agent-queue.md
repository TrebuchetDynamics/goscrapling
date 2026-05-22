# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Static browser impersonation, HTTP/3, and stealthy headers boundary

- Phase: `phase-1-response-fetcher / static-fetcher`
- Priority: `P2`
- Owner: `fetcher`
- Size: `medium`
- Contract status: `draft`
- Contract: Evaluate and implement the Go-native boundary for Scrapling static fetcher impersonation claims, including browser-like headers, TLS/browser impersonation, and HTTP/3 where supported.
- Ready when: Core static options and proxy support are complete.
- Write scope: `fetcher_identity.go`, `fetcher_identity_test.go`, `docs/content/building-goscrapling/architecture_plan/boundaries.md`
- Test commands: `go test ./... -run TestStaticFetcherIdentityOptions -count=1`
- Acceptance: Fixtures prove explicit header generation and any unsupported impersonation modes return honest operator-visible errors.
- Done signal: Static identity option tests pass and boundaries document supported claims.
- Source refs: `references/Scrapling/scrapling/engines/toolbelt/fingerprints.py`, `references/Scrapling/scrapling/fetchers/requests.py`, `references/Scrapling/docs/fetching/static.md`
<!-- PROGRESS:END -->
