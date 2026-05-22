# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Proxy rotator strategies

- Phase: `phase-1-response-fetcher / static-fetcher`
- Priority: `P1`
- Owner: `fetcher`
- Size: `medium`
- Contract status: `draft`
- Contract: Port ProxyRotator behavior with cyclic rotation, custom strategy hooks, string/dictionary-style configuration mapping, and proxy-error retry integration.
- Ready when: Static proxy support is complete.
- Write scope: `proxy.go`, `proxy_rotator.go`, `proxy_rotator_test.go`
- Test commands: `go test ./... -run TestProxyRotator -count=1`
- Acceptance: Unit fixtures prove cyclic/custom rotation, exhausted proxy behavior, proxy config parsing, and retry-on-proxy-error semantics.
- Done signal: Proxy rotator tests pass.
- Source refs: `references/Scrapling/scrapling/engines/toolbelt/proxy_rotation.py`, `references/Scrapling/docs/api-reference/proxy-rotation.md`, `references/Scrapling/docs/spiders/proxy-blocking.md`
<!-- PROGRESS:END -->
