# Scrapling Behavior Atoms

This file records the smallest meaningful Scrapling behaviors that goscrapling
must preserve or intentionally replace. It is not a side backlog. Every
incomplete atom must point at a `progress.json` row.

Status vocabulary: `covered`, `partial`, `planned`, `vague`, `owned`,
`excluded`.

## Response And Static Fetcher Atoms

| Atom | Upstream refs | Visible contract | Go target | Progress row | Validation | Status |
|---|---|---|---|---|---|---|
| Response metadata and selector behavior | `references/Scrapling/docs/api-reference/response.md`, `references/Scrapling/docs/fetching/choosing.md`, `references/Scrapling/scrapling/engines/static.py` | A fetched or constructed response exposes URL, status, headers, body metadata, and selector methods on the parsed document. | `Response`, `Document`, `Element`, `Selection` | `Response metadata and selector contract` | `go test ./... -run TestResponseMetadataAndSelectorContract -count=1` | planned |
| Response body, text, bytes, and JSON helpers | `references/Scrapling/docs/api-reference/response.md`, `references/Scrapling/docs/fetching/static.md`, `references/Scrapling/scrapling/engines/toolbelt/custom.py` | Response consumers can read immutable body bytes, decoded text, and JSON values with visible decode errors. | `Response` | `Response body, text, bytes, and JSON helpers` | `go test ./... -run TestResponseBodyHelpers -count=1` | planned |
| Static Fetcher methods | `references/Scrapling/docs/fetching/static.md`, `references/Scrapling/scrapling/fetchers/requests.py` | Static fetcher exposes GET, POST, PUT, and DELETE request paths that return Response values. | `Fetcher` | `Static Fetcher method surface over net/http` | `go test ./... -run TestStaticFetcherMethods -count=1` | planned |
| FetcherSession defaults and cookies | `references/Scrapling/docs/fetching/static.md`, `references/Scrapling/scrapling/engines/static.py` | A session merges defaults with per-request options and preserves cookies across requests. | `FetcherSession` | `FetcherSession option merging and cookies` | `go test ./... -run TestFetcherSessionOptionMergingAndCookies -count=1` | planned |
| Redirect, timeout, retry, and error taxonomy | `references/Scrapling/docs/fetching/static.md`, `references/Scrapling/scrapling/engines/static.py`, `references/Scrapling/scrapling/engines/toolbelt/navigation.py` | Static fetching classifies safe redirects, disabled redirects, private-address redirects, timeouts, and retry exhaustion deterministically. | `Fetcher`, `FetcherError` | `Redirect, timeout, retry, and error taxonomy` | `go test ./... -run TestFetcherRedirectTimeoutRetryErrors -count=1` | planned |

## Browser Fetching Atoms

| Atom | Upstream refs | Visible contract | Go target | Progress row | Validation | Status |
|---|---|---|---|---|---|---|
| Browser fetcher contract | `references/Scrapling/docs/fetching/dynamic.md`, `references/Scrapling/scrapling/fetchers/chrome.py`, `references/Scrapling/scrapling/engines/_browsers/_page.py` | Dynamic fetcher can navigate, wait, optionally run page actions, and return a Response without exposing engine internals. | `browser` package | `BrowserFetcher interface and chromedp/playwright adapter decision` | `go test ./... -run TestBrowserFetcherContract -count=1` | planned |
| Stealth browser controls | `references/Scrapling/docs/fetching/stealthy.md`, `references/Scrapling/scrapling/fetchers/stealth_chrome.py`, `references/Scrapling/scrapling/engines/_browsers/_stealth.py` | Stealth behavior is explicit, test-backed, and separate from normal browser fetching claims. | future `browser` options | future split under Phase 3 | browser fixture plus explicit operator docs | vague |

## Spider And Crawler Atoms

| Atom | Upstream refs | Visible contract | Go target | Progress row | Validation | Status |
|---|---|---|---|---|---|---|
| Spider request object | `references/Scrapling/docs/spiders/requests-responses.md`, `references/Scrapling/scrapling/spiders/request.py` | Request carries URL, method/body metadata, session id, priority, dedupe flag, callback, and meta. | `spider.Request` | `Spider request, result, scheduler, and session contracts` | `go test ./... -run TestSpider -count=1` | planned |
| Response follow behavior | `references/Scrapling/docs/spiders/requests-responses.md` | `Response.follow` resolves relative URLs, inherits request context, and applies referer flow unless disabled. | `spider.Response` or `Response.Follow` | `Spider request, result, scheduler, and session contracts` | `go test ./... -run TestSpider -count=1` | planned |
| Priority scheduler and dedupe | `references/Scrapling/docs/spiders/architecture.md`, `references/Scrapling/scrapling/spiders/scheduler.py` | Scheduler dequeues higher priority first and deduplicates by request fingerprint unless disabled. | `spider.Scheduler` | `Spider request, result, scheduler, and session contracts` | `go test ./... -run TestSpider -count=1` | planned |
| Session routing | `references/Scrapling/docs/spiders/sessions.md`, `references/Scrapling/scrapling/spiders/session.py` | Session manager routes requests by session id and supports lazy or eager session startup. | `spider.SessionManager` | `Spider request, result, scheduler, and session contracts` | `go test ./... -run TestSpider -count=1` | planned |
| Crawl result and stats | `references/Scrapling/docs/spiders/architecture.md`, `references/Scrapling/scrapling/spiders/result.py` | Crawl output records items, errors, skipped requests, and crawl statistics. | `spider.Result`, `spider.Stats` | `Spider request, result, scheduler, and session contracts` | `go test ./... -run TestSpider -count=1` | planned |
| Robots, cache, checkpoint, and stats controls | `references/Scrapling/scrapling/spiders/robotstxt.py`, `references/Scrapling/scrapling/spiders/cache.py`, `references/Scrapling/scrapling/spiders/checkpoint.py` | Production crawling controls are split into independent rows after spider core is proven. | `spider` package | `Robots, cache, checkpoint, and stats as separate crawler slices` | split-specific fixture tests | vague |

## Parser And Adaptive Atoms Already Covered

| Atom | Upstream refs | Visible contract | Go target | Progress row | Validation | Status |
|---|---|---|---|---|---|---|
| Static HTML parse and CSS selection | `references/Scrapling/scrapling/parser.py` | Parse HTML and query elements by CSS selectors. | `Document`, `Element`, `Selection` | `Parse static HTML into queryable documents` | `go test ./... -run TestParseSelectsElementsWithCSS -count=1` | covered |
| Adaptive fingerprint save | `references/Scrapling/scrapling/core/storage.py` | Save fingerprints under domain and identifier. | `AdaptiveStore`, `MemoryStore` | `Save adaptive fingerprints by domain and identifier` | `go test ./... -run 'TestSaveStoresFingerprintByDomainAndIdentifier|TestSaveIsolatesFingerprintsByDomain|TestParseUsesDefaultDomainWhenURLIsMissing' -count=1` | covered |
| Adaptive relocation | `references/Scrapling/scrapling/parser.py` | Relocate a logical element after markup changes without accepting low-confidence matches. | `Document.Relocate` | `Relocate adaptive elements with deterministic scoring` | `go test ./... -run TestRelocate -count=1` | covered |
