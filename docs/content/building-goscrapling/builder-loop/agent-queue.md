# Agent Queue

This page is generated from canonical `progress.json` rows that are unblocked,
non-umbrella, and builder-ready.

<!-- PROGRESS:START kind=agent-queue -->
## 1. Async and concurrent static fetcher API

- Phase: `phase-1-response-fetcher / static-fetcher`
- Priority: `P2`
- Owner: `fetcher`
- Size: `medium`
- Contract status: `draft`
- Contract: Provide a Go-native concurrent equivalent for Scrapling AsyncFetcher and async sessions using context, goroutines, cancellation, and bounded concurrency.
- Ready when: Static fetcher options are stable.
- Write scope: `fetchers/fetcher_async.go`, `fetchers/fetcher_async_test.go`
- Test commands: `go test ./... -run TestConcurrentFetcher -count=1`
- Acceptance: httptest fixtures prove concurrent fetch, cancellation, error collection, and session reuse without Python-style async mimicry.
- Done signal: Concurrent fetcher tests pass.
- Source refs: `references/Scrapling/scrapling/fetchers/requests.py`, `references/Scrapling/docs/api-reference/fetchers.md`
<!-- PROGRESS:END -->
