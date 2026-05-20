# Blocked Slices

This page is generated from canonical `progress.json` rows that declare
`blocked_by`.

<!-- PROGRESS:START kind=blocked-slices -->
| Phase | Slice | Blocked by | Ready when | Unblocks |
|---|---|---|---|---|
| phase-1-response-fetcher / static-fetcher | Proxy rotator strategies | Static proxy support and proxy error classification | Static proxy support is complete. | - |
| phase-3-browser / browser-fetcher | Browser context options and resource blocking | Browser sessions and page pool lifecycle | Browser adapter and session rows are stable. | - |
| phase-3-browser / browser-fetcher | Browser sessions and page pool lifecycle | Real browser adapter with JavaScript fixture | Real browser adapter or fake-backed engine seam is stable. | - |
| phase-3-browser / browser-fetcher | Browser wait conditions, page actions, downloads, screenshots, and XHR capture | Browser sessions and page pool lifecycle | Browser session pool is stable. | - |
| phase-4-spider / spider-core | Blocked response detection and retry hooks | Crawler engine concurrency, domain limits, and download delay | Crawler engine loop and stats are stable. | - |
| phase-4-spider / spider-core | Checkpoint pause and resume | Crawler engine concurrency, domain limits, and download delay | Scheduler fingerprints and crawler loop are stable. | - |
| phase-4-spider / spider-core | Robots.txt manager and delay directives | Crawler engine concurrency, domain limits, and download delay | Crawler engine filtering hooks are available. | - |
| phase-4-spider / spider-core | Spider lifecycle hooks, streaming, item hooks, and expanded stats | Crawler engine concurrency, domain limits, and download delay | Crawler engine loop and result struct are stable. | - |
| phase-0-parser-adaptive / static-document | Selector generation helpers | Traversal, filtering, and text search helpers | Traversal helpers expose enough DOM context to generate selectors. | - |
| phase-1-response-fetcher / static-fetcher | Async and concurrent static fetcher API | Static request options: params, forms, JSON, auth, verify, and cookies | Static fetcher options are stable. | - |
| phase-1-response-fetcher / static-fetcher | Static browser impersonation, HTTP/3, and stealthy headers boundary | Static request options: params, forms, JSON, auth, verify, and cookies | Core static options and proxy support are complete. | - |
| phase-3-browser / browser-fetcher | Stealth browser controls and fingerprint options | Browser context options and resource blocking | Normal browser context options are stable. | - |
| phase-4-spider / spider-core | LinkExtractor and crawl templates | XPath selection and CSS-to-XPath translator parity | Response.follow and selector helpers are stable. | - |
| phase-4-spider / spider-core | Spider static, dynamic, stealth, and proxy session adapters | Proxy rotator strategies, Browser sessions and page pool lifecycle, Stealth browser controls and fingerprint options | Fetcher, browser session, and proxy rows are complete. | - |
| phase-5-cli-tooling / tool-surfaces | MCP scraping tool server | Browser wait conditions, page actions, downloads, screenshots, and XHR capture, Stealth browser controls and fingerprint options | Static CLI extract and fetcher APIs are stable enough for tool output fixtures. | - |
| phase-3-browser / browser-fetcher | Cloudflare challenge strategy boundary | Stealth browser controls and fingerprint options | Stealth browser controls are explicit and tested. | - |
| phase-5-cli-tooling / tool-surfaces | CLI extract markdown, AI-targeted, and browser modes | Browser wait conditions, page actions, downloads, screenshots, and XHR capture, Stealth browser controls and fingerprint options | Static extract methods are complete, Browser fetcher has a real adapter or stable fake-backed command seam. | - |
| phase-5-cli-tooling / tool-surfaces | Install command, Docker image, and dependency packaging docs | Real browser adapter with JavaScript fixture | Browser adapter dependency boundary is documented. | - |
| phase-5-cli-tooling / tool-surfaces | Benchmarks and parity scorecard | Public docs, examples, and API reference parity | Major subsystem APIs are stable enough to benchmark. | - |
<!-- PROGRESS:END -->
