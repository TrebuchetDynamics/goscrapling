# Next Slices

This page is generated from the canonical progress file and lists the highest
leverage rows to execute next.

<!-- PROGRESS:START kind=next-slices -->
| Phase | Slice | Owner | Size | Contract status | Why now |
|---|---|---|---|---|---|
| phase-1-response-fetcher / response | Response metadata and selector contract | `fetcher` | `small` | `fixture_ready` | P0 port parity row with complete handoff metadata. |
| phase-2-storage / persistent-store | File-backed adaptive store with compatibility migration | `storage` | `medium` | `draft` | Contract metadata is present and the row is unblocked. |
| phase-3-browser / browser-fetcher | BrowserFetcher interface and chromedp/playwright adapter decision | `browser` | `medium` | `draft` | Contract metadata is present and the row is unblocked. |
| phase-4-spider / spider-core | Spider request, result, scheduler, and session contracts | `spider` | `large` | `draft` | Contract metadata is present and the row is unblocked. |
| phase-5-cli-tooling / tool-surfaces | CLI extraction workflow parity | `cli` | `medium` | `draft` | Contract metadata is present and the row is unblocked. |
| phase-5-cli-tooling / tool-surfaces | Gormes web-search tool adapter | `integration` | `medium` | `draft` | Contract metadata is present and the row is unblocked. |
<!-- PROGRESS:END -->
