# Upstream Scrapling App Map

## Source Metadata

- Version: `1.0`
- Upstream name: `D4Vinci/Scrapling`
- Upstream repo: `https://github.com/D4Vinci/Scrapling`
- Observed commit: `6380ef0f266a5fff898c18953d6b03ca320b2fd4`
- Observed release: `v0.4.8`
- Local checkout: `references/Scrapling`
- Generated markdown: `docs/content/building-goscrapling/architecture_plan/upstream-app-map.md`
- py2many probe dir: `docs/research/python-to-go-probes/py2many`

## Coverage Summary

| Status | Count |
|---|---:|
| `partial` | 11 |
| `planned` | 9 |
| `excluded` | 3 |

## Entries

| Entry | Status | Feature anchor | Go target | Progress rows | Translation suitability | Upstream refs |
|---|---|---|---|---|---|---:|
| Package Metadata And Public Exports | `partial` | `Package metadata and public exports` | `root package exports plus future package-level adapters` | `Parse static HTML into queryable documents`, `Response metadata and selector contract`, `Static Fetcher method surface over net/http`, `BrowserFetcher interface and chromedp/playwright adapter decision`, `Spider request, result, scheduler, and session contracts`, `CLI static extract GET command` | `manual_rewrite` | 9 |
| Python Typing Marker | `excluded` | `-` | `-` | - | `not_applicable` | 1 |
| Parser, Selectors, Custom Types, And Translator | `partial` | `Parser and selector objects` | `Document, Element, Selection, selector helper types, and future translator helpers` | `Parse static HTML into queryable documents`, `Selector pseudo-elements, extraction helpers, and custom types`, `XPath selection and CSS-to-XPath translator parity`, `Traversal, filtering, and text search helpers`, `Selector generation helpers` | `manual_rewrite` | 5 |
| Adaptive Storage And Relocation Helpers | `partial` | `Adaptive storage` | `AdaptiveStore, MemoryStore, FileStore, SQLiteStore, and relocation scoring` | `Save adaptive fingerprints by domain and identifier`, `Relocate adaptive elements with deterministic scoring`, `Adaptive selector fallback and auto-save modes`, `Adaptive fingerprint parity review and diagnostics`, `File-backed adaptive store with compatibility migration`, `SQLite adaptive store parity` | `manual_rewrite` | 2 |
| Static Response And Fetchers | `partial` | `Static fetcher` | `Response, Fetcher, FetcherSession, static request options, and response conversion helpers` | `Response metadata and selector contract`, `Response body, text, bytes, and JSON helpers`, `Response cookies, history, meta, and captured XHR fields`, `Static Fetcher method surface over net/http`, `FetcherSession option merging and cookies`, `Redirect, timeout, retry, and error taxonomy`, `Static request options: params, forms, JSON, auth, verify, and cookies`, `Async and concurrent static fetcher API` | `manual_rewrite` | 4 |
| Proxy Rotation, Impersonation, And Network Helper Constants | `planned` | `Proxy rotation` | `ProxyRotator, static proxy support, proxy error classification, and explicit impersonation boundaries` | `Static proxy support and proxy error classification`, `Proxy rotator strategies`, `Static browser impersonation, HTTP/3, and stealthy headers boundary`, `Spider static, dynamic, stealth, and proxy session adapters` | `manual_rewrite` | 2 |
| Browser And Stealth Fetchers | `partial` | `Browser fetching` | `BrowserFetcher, ChromedpBrowserEngine, future browser sessions, page pools, context options, and stealth option structs` | `BrowserFetcher interface and chromedp/playwright adapter decision`, `Real browser adapter with JavaScript fixture`, `Browser sessions and page pool lifecycle`, `Browser wait conditions, page actions, downloads, screenshots, and XHR capture`, `Browser context options and resource blocking`, `Stealth browser controls and fingerprint options`, `Cloudflare challenge strategy boundary` | `manual_rewrite` | 12 |
| Spider Core And Crawler Engine | `partial` | `Spider runtime` | `spider package request/result/session/scheduler types, allowed-domain filtering, and concurrent crawler engine controls` | `Spider request, result, scheduler, and session contracts`, `Crawler engine concurrency, domain limits, and download delay`, `Allowed domains and offsite filtering`, `Blocked response detection and retry hooks`, `Spider lifecycle hooks, streaming, item hooks, and expanded stats`, `Spider static, dynamic, stealth, and proxy session adapters` | `manual_rewrite` | 6 |
| Robots, Cache, Checkpoint, And Production Crawler Controls | `planned` | `Spider runtime` | `spider package robots manager, response cache, checkpoint store, and crawler controls` | `Robots.txt manager and delay directives`, `Development response cache`, `Checkpoint pause and resume` | `probe_candidate` | 3 |
| Link Extraction And Spider Templates | `planned` | `Spider runtime` | `spider package LinkExtractor plus crawler and sitemap templates` | `LinkExtractor and crawl templates` | `probe_candidate` | 3 |
| CLI Extract And Shell | `partial` | `CLI shell and extract commands` | `cmd/goscrapling and internal/cli` | `CLI static extract GET command`, `CLI static extract method and request body expansion`, `CLI extract markdown, AI-targeted, and browser modes`, `CLI interactive shell command surface`, `Install command, Docker image, and dependency packaging docs` | `manual_rewrite` | 4 |
| MCP And AI Tools | `planned` | `MCP and AI docs` | `integrations/mcp and future integration packages` | `MCP scraping tool server` | `manual_rewrite` | 1 |
| Upstream Parser And Adaptive Tests | `partial` | `Parser and selector objects` | `parser and adaptive test fixtures in root package` | `Parse static HTML into queryable documents`, `Selector pseudo-elements, extraction helpers, and custom types`, `XPath selection and CSS-to-XPath translator parity`, `Traversal, filtering, and text search helpers`, `Selector generation helpers`, `Relocate adaptive elements with deterministic scoring` | `not_applicable` | 8 |
| Upstream Fetcher And Browser Tests | `partial` | `Static fetcher` | `fetcher, response, proxy, browser, and fake-browser fixture tests` | `Response metadata and selector contract`, `Response body, text, bytes, and JSON helpers`, `Static Fetcher method surface over net/http`, `FetcherSession option merging and cookies`, `Redirect, timeout, retry, and error taxonomy`, `Static proxy support and proxy error classification`, `Proxy rotator strategies`, `BrowserFetcher interface and chromedp/playwright adapter decision`, `Stealth browser controls and fingerprint options` | `not_applicable` | 22 |
| Upstream Spider Tests | `partial` | `Spider runtime` | `spider package fake-fetcher and scheduler fixtures` | `Spider request, result, scheduler, and session contracts`, `Crawler engine concurrency, domain limits, and download delay`, `Robots.txt manager and delay directives`, `Development response cache`, `Checkpoint pause and resume`, `LinkExtractor and crawl templates` | `not_applicable` | 14 |
| Upstream CLI, Core, Storage, And AI Tests | `partial` | `CLI shell and extract commands` | `internal/cli, adaptive storage tests, and integrations/mcp fixtures` | `CLI static extract GET command`, `CLI static extract method and request body expansion`, `CLI interactive shell command surface`, `Save adaptive fingerprints by domain and identifier`, `File-backed adaptive store with compatibility migration`, `MCP scraping tool server` | `not_applicable` | 9 |
| Upstream Parser And Adaptive Docs | `planned` | `Parser and selector objects` | `root parser/adaptive docs, examples, and API reference` | `Selector pseudo-elements, extraction helpers, and custom types`, `XPath selection and CSS-to-XPath translator parity`, `Adaptive selector fallback and auto-save modes`, `Public docs, examples, and API reference parity` | `not_applicable` | 7 |
| Upstream Response, Fetching, Proxy, And Browser Docs | `planned` | `Static fetcher` | `Response, Fetcher, FetcherSession, ProxyRotator, and browser docs` | `Response cookies, history, meta, and captured XHR fields`, `Static request options: params, forms, JSON, auth, verify, and cookies`, `Proxy rotator strategies`, `Real browser adapter with JavaScript fixture`, `Stealth browser controls and fingerprint options`, `Public docs, examples, and API reference parity` | `not_applicable` | 7 |
| Upstream Spider Docs | `planned` | `Spider runtime` | `spider package docs, examples, and templates` | `Spider request, result, scheduler, and session contracts`, `Crawler engine concurrency, domain limits, and download delay`, `Robots.txt manager and delay directives`, `LinkExtractor and crawl templates`, `Spider static, dynamic, stealth, and proxy session adapters`, `Public docs, examples, and API reference parity` | `not_applicable` | 8 |
| Upstream CLI, MCP, And AI Docs | `planned` | `CLI shell and extract commands` | `cmd/goscrapling, internal/cli, and integrations/mcp docs` | `CLI static extract GET command`, `CLI static extract method and request body expansion`, `CLI extract markdown, AI-targeted, and browser modes`, `CLI interactive shell command surface`, `MCP scraping tool server`, `Public docs, examples, and API reference parity` | `not_applicable` | 5 |
| Upstream Public Docs, Tutorials, Donate Page, And Benchmarks | `planned` | `Install, Docker, packaging, examples, and benchmarks` | `README.md, docs, example tests, benchmarks, and generated scorecards` | `Public docs, examples, and API reference parity`, `Benchmarks and parity scorecard`, `Install command, Docker image, and dependency packaging docs` | `not_applicable` | 6 |
| Translated README Files | `excluded` | `-` | `-` | - | `not_applicable` | 9 |
| Docs Assets And Branding Images | `excluded` | `-` | `-` | - | `not_applicable` | 9 |

## Package Metadata And Public Exports

- ID: `01-package-public-exports`
- Upstream refs:
  - `references/Scrapling/scrapling/__init__.py` (`source`)
  - `references/Scrapling/scrapling/core/__init__.py` (`source`)
  - `references/Scrapling/scrapling/core/utils/__init__.py` (`source`)
  - `references/Scrapling/scrapling/engines/__init__.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/__init__.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/__init__.py` (`source`)
  - `references/Scrapling/scrapling/fetchers/__init__.py` (`source`)
  - `references/Scrapling/scrapling/spiders/__init__.py` (`source`)
  - `references/Scrapling/scrapling/spiders/templates/__init__.py` (`source`)
- Behavior atoms: Export Scrapling-visible parser, fetcher, browser, spider, CLI, and tool symbols through stable package surfaces
- Notes: Python module export wiring is a parity signal, but Go package shape remains intentionally Go-native.

## Python Typing Marker

- ID: `02-python-typing-marker`
- Upstream refs:
  - `references/Scrapling/scrapling/py.typed` (`source`)
- Notes: Python packaging type marker has no Go runtime or docs parity requirement.

## Parser, Selectors, Custom Types, And Translator

- ID: `03-parser-selectors-custom-types`
- Upstream refs:
  - `references/Scrapling/scrapling/core/_types.py` (`source`)
  - `references/Scrapling/scrapling/core/custom_types.py` (`source`)
  - `references/Scrapling/scrapling/core/mixins.py` (`source`)
  - `references/Scrapling/scrapling/core/translator.py` (`source`)
  - `references/Scrapling/scrapling/parser.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/parser.go.txt`, `docs/research/python-to-go-probes/py2many/go/core__custom_types.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Parse HTML into queryable documents and element selections, Preserve selector helpers, pseudo-elements, extraction handlers, traversal, filtering, and XPath/CSS translation behavior
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## Adaptive Storage And Relocation Helpers

- ID: `04-adaptive-storage`
- Upstream refs:
  - `references/Scrapling/scrapling/core/storage.py` (`source`)
  - `references/Scrapling/scrapling/core/utils/_utils.py` (`source`)
- Behavior atoms: Store adaptive fingerprints by domain and identifier, Relocate elements through deterministic fingerprint scoring and expose adaptive selector fallback modes

## Static Response And Fetchers

- ID: `05-response-static-fetchers`
- Upstream refs:
  - `references/Scrapling/scrapling/engines/static.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/convertor.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/custom.py` (`source`)
  - `references/Scrapling/scrapling/fetchers/requests.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/engines__static.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Fetch static responses with requests-style options, sessions, redirects, retries, and response conversion helpers, Expose response metadata, body helpers, selector access, cookies, history, meta, and captured XHR attachment points
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## Proxy Rotation, Impersonation, And Network Helper Constants

- ID: `06-proxy-impersonation-network-helpers`
- Upstream refs:
  - `references/Scrapling/scrapling/engines/constants.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/proxy_rotation.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/engines__toolbelt__proxy_rotation.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Represent proxy rotation strategies and network error classifications as explicit fetcher/spider options, Keep browser impersonation and HTTP/3 behavior bounded by testable Go dependencies
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## Browser And Stealth Fetchers

- ID: `07-browser-stealth-fetchers`
- Upstream refs:
  - `references/Scrapling/scrapling/engines/_browsers/_base.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/_config_tools.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/_controllers.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/_page.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/_stealth.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/_types.py` (`source`)
  - `references/Scrapling/scrapling/engines/_browsers/_validators.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/ad_domains.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/fingerprints.py` (`source`)
  - `references/Scrapling/scrapling/engines/toolbelt/navigation.py` (`source`)
  - `references/Scrapling/scrapling/fetchers/chrome.py` (`source`)
  - `references/Scrapling/scrapling/fetchers/stealth_chrome.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/engines__toolbelt__navigation.go.txt`, `docs/research/python-to-go-probes/py2many/go/engines___browsers___validators.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Map browser fetcher lifecycle, page operations, waits, downloads, screenshots, XHR capture, context options, resource blocking, and stealth options to explicit Go adapters, Keep stealth and challenge handling operator-visible and test-backed
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## Spider Core And Crawler Engine

- ID: `08-spider-core-crawler-engine`
- Upstream refs:
  - `references/Scrapling/scrapling/spiders/engine.py` (`source`)
  - `references/Scrapling/scrapling/spiders/request.py` (`source`)
  - `references/Scrapling/scrapling/spiders/result.py` (`source`)
  - `references/Scrapling/scrapling/spiders/scheduler.py` (`source`)
  - `references/Scrapling/scrapling/spiders/session.py` (`source`)
  - `references/Scrapling/scrapling/spiders/spider.py` (`source`)
- Behavior atoms: Model spider request/result/session/scheduler contracts and crawler engine execution semantics, Preserve concurrency, domain limits, download delays, lifecycle hooks, streaming, item hooks, stats, and session adapter boundaries

## Robots, Cache, Checkpoint, And Production Crawler Controls

- ID: `09-spider-production-controls`
- Upstream refs:
  - `references/Scrapling/scrapling/spiders/cache.py` (`source`)
  - `references/Scrapling/scrapling/spiders/checkpoint.py` (`source`)
  - `references/Scrapling/scrapling/spiders/robotstxt.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/spiders__robotstxt.go.txt`, `docs/research/python-to-go-probes/py2many/go/spiders__cache.go.txt`, `docs/research/python-to-go-probes/py2many/go/spiders__checkpoint.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Preserve robots.txt decisions, crawl-delay handling, development response cache behavior, and checkpoint pause/resume state
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## Link Extraction And Spider Templates

- ID: `10-link-extraction-spider-templates`
- Upstream refs:
  - `references/Scrapling/scrapling/spiders/links.py` (`source`)
  - `references/Scrapling/scrapling/spiders/templates/crawler.py` (`source`)
  - `references/Scrapling/scrapling/spiders/templates/sitemap.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/spiders__links.go.txt`, `docs/research/python-to-go-probes/py2many/go/spiders__templates__crawler.go.txt`, `docs/research/python-to-go-probes/py2many/go/spiders__templates__sitemap.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Extract and normalize links, then expose crawler and sitemap templates as higher-level spider building blocks
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## CLI Extract And Shell

- ID: `11-cli-extract-shell`
- Upstream refs:
  - `references/Scrapling/scrapling/cli.py` (`source`)
  - `references/Scrapling/scrapling/core/_shell_signatures.py` (`source`)
  - `references/Scrapling/scrapling/core/shell.py` (`source`)
  - `references/Scrapling/scrapling/core/utils/_shell.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/cli.go.txt`, `docs/research/python-to-go-probes/py2many/go/core__shell.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Expose extract commands for static, browser, stealth, markdown, and AI-targeted modes, Map Scrapling shell shortcuts and interactive concepts into scripted Go command fixtures before a full REPL
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## MCP And AI Tools

- ID: `12-mcp-ai-tools`
- Upstream refs:
  - `references/Scrapling/scrapling/core/ai.py` (`source`)
- Static reference paths: `docs/research/python-to-go-probes/py2many/go/core__ai.go.txt`, `docs/research/python-to-go-probes/py2many/summary.md`
- Behavior atoms: Expose deterministic MCP scraping tool schemas and no-live-LLM integration boundaries
- Notes: The current py2many artifacts are missing-tool/reference-only evidence and not generated Go implementation.

## Upstream Parser And Adaptive Tests

- ID: `13-tests-parser-adaptive`
- Upstream refs:
  - `references/Scrapling/tests/parser/__init__.py` (`test`)
  - `references/Scrapling/tests/parser/test_adaptive.py` (`test`)
  - `references/Scrapling/tests/parser/test_ancestor_navigation.py` (`test`)
  - `references/Scrapling/tests/parser/test_attributes_handler.py` (`test`)
  - `references/Scrapling/tests/parser/test_find_similar_advanced.py` (`test`)
  - `references/Scrapling/tests/parser/test_general.py` (`test`)
  - `references/Scrapling/tests/parser/test_parser_advanced.py` (`test`)
  - `references/Scrapling/tests/parser/test_selectors_filter.py` (`test`)

## Upstream Fetcher And Browser Tests

- ID: `14-tests-fetchers-browser`
- Upstream refs:
  - `references/Scrapling/tests/fetchers/__init__.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/__init__.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/test_dynamic.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/test_dynamic_session.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/test_requests.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/test_requests_session.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/test_stealth.py` (`test`)
  - `references/Scrapling/tests/fetchers/async/test_stealth_session.py` (`test`)
  - `references/Scrapling/tests/fetchers/sync/__init__.py` (`test`)
  - `references/Scrapling/tests/fetchers/sync/test_dynamic.py` (`test`)
  - `references/Scrapling/tests/fetchers/sync/test_requests.py` (`test`)
  - `references/Scrapling/tests/fetchers/sync/test_requests_session.py` (`test`)
  - `references/Scrapling/tests/fetchers/sync/test_stealth_session.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_base.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_constants.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_impersonate_list.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_merge_request_args.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_pages.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_proxy_rotation.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_response_handling.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_utils.py` (`test`)
  - `references/Scrapling/tests/fetchers/test_validator.py` (`test`)

## Upstream Spider Tests

- ID: `15-tests-spiders`
- Upstream refs:
  - `references/Scrapling/tests/spiders/__init__.py` (`test`)
  - `references/Scrapling/tests/spiders/test_cache.py` (`test`)
  - `references/Scrapling/tests/spiders/test_checkpoint.py` (`test`)
  - `references/Scrapling/tests/spiders/test_engine.py` (`test`)
  - `references/Scrapling/tests/spiders/test_force_stop_checkpoint.py` (`test`)
  - `references/Scrapling/tests/spiders/test_links.py` (`test`)
  - `references/Scrapling/tests/spiders/test_request.py` (`test`)
  - `references/Scrapling/tests/spiders/test_result.py` (`test`)
  - `references/Scrapling/tests/spiders/test_robotstxt.py` (`test`)
  - `references/Scrapling/tests/spiders/test_scheduler.py` (`test`)
  - `references/Scrapling/tests/spiders/test_session.py` (`test`)
  - `references/Scrapling/tests/spiders/test_sitemap.py` (`test`)
  - `references/Scrapling/tests/spiders/test_spider.py` (`test`)
  - `references/Scrapling/tests/spiders/test_templates.py` (`test`)

## Upstream CLI, Core, Storage, And AI Tests

- ID: `16-tests-cli-core-ai`
- Upstream refs:
  - `references/Scrapling/tests/__init__.py` (`test`)
  - `references/Scrapling/tests/ai/__init__.py` (`test`)
  - `references/Scrapling/tests/ai/test_ai_mcp.py` (`test`)
  - `references/Scrapling/tests/cli/__init__.py` (`test`)
  - `references/Scrapling/tests/cli/test_cli.py` (`test`)
  - `references/Scrapling/tests/cli/test_shell_functionality.py` (`test`)
  - `references/Scrapling/tests/core/__init__.py` (`test`)
  - `references/Scrapling/tests/core/test_shell_core.py` (`test`)
  - `references/Scrapling/tests/core/test_storage_core.py` (`test`)

## Upstream Parser And Adaptive Docs

- ID: `17-docs-parser-adaptive`
- Upstream refs:
  - `references/Scrapling/docs/api-reference/custom-types.md` (`doc`)
  - `references/Scrapling/docs/api-reference/selector.md` (`doc`)
  - `references/Scrapling/docs/development/adaptive_storage_system.md` (`doc`)
  - `references/Scrapling/docs/development/scrapling_custom_types.md` (`doc`)
  - `references/Scrapling/docs/parsing/adaptive.md` (`doc`)
  - `references/Scrapling/docs/parsing/main_classes.md` (`doc`)
  - `references/Scrapling/docs/parsing/selection.md` (`doc`)

## Upstream Response, Fetching, Proxy, And Browser Docs

- ID: `18-docs-response-fetching-browser`
- Upstream refs:
  - `references/Scrapling/docs/api-reference/fetchers.md` (`doc`)
  - `references/Scrapling/docs/api-reference/proxy-rotation.md` (`doc`)
  - `references/Scrapling/docs/api-reference/response.md` (`doc`)
  - `references/Scrapling/docs/fetching/choosing.md` (`doc`)
  - `references/Scrapling/docs/fetching/dynamic.md` (`doc`)
  - `references/Scrapling/docs/fetching/static.md` (`doc`)
  - `references/Scrapling/docs/fetching/stealthy.md` (`doc`)

## Upstream Spider Docs

- ID: `19-docs-spider-runtime`
- Upstream refs:
  - `references/Scrapling/docs/api-reference/spiders.md` (`doc`)
  - `references/Scrapling/docs/spiders/advanced.md` (`doc`)
  - `references/Scrapling/docs/spiders/architecture.md` (`doc`)
  - `references/Scrapling/docs/spiders/generic-templates.md` (`doc`)
  - `references/Scrapling/docs/spiders/getting-started.md` (`doc`)
  - `references/Scrapling/docs/spiders/proxy-blocking.md` (`doc`)
  - `references/Scrapling/docs/spiders/requests-responses.md` (`doc`)
  - `references/Scrapling/docs/spiders/sessions.md` (`doc`)

## Upstream CLI, MCP, And AI Docs

- ID: `20-docs-cli-mcp-ai`
- Upstream refs:
  - `references/Scrapling/docs/ai/mcp-server.md` (`doc`)
  - `references/Scrapling/docs/api-reference/mcp-server.md` (`doc`)
  - `references/Scrapling/docs/cli/extract-commands.md` (`doc`)
  - `references/Scrapling/docs/cli/interactive-shell.md` (`doc`)
  - `references/Scrapling/docs/cli/overview.md` (`doc`)

## Upstream Public Docs, Tutorials, Donate Page, And Benchmarks

- ID: `21-docs-public-examples-benchmarks`
- Upstream refs:
  - `references/Scrapling/docs/benchmarks.md` (`doc`)
  - `references/Scrapling/docs/donate.md` (`doc`)
  - `references/Scrapling/docs/index.md` (`doc`)
  - `references/Scrapling/docs/overview.md` (`doc`)
  - `references/Scrapling/docs/tutorials/migrating_from_beautifulsoup.md` (`doc`)
  - `references/Scrapling/docs/tutorials/replacing_ai.md` (`doc`)

## Translated README Files

- ID: `22-docs-translated-readmes`
- Upstream refs:
  - `references/Scrapling/docs/README_AR.md` (`doc`)
  - `references/Scrapling/docs/README_CN.md` (`doc`)
  - `references/Scrapling/docs/README_DE.md` (`doc`)
  - `references/Scrapling/docs/README_ES.md` (`doc`)
  - `references/Scrapling/docs/README_FR.md` (`doc`)
  - `references/Scrapling/docs/README_JP.md` (`doc`)
  - `references/Scrapling/docs/README_KR.md` (`doc`)
  - `references/Scrapling/docs/README_PT_BR.md` (`doc`)
  - `references/Scrapling/docs/README_RU.md` (`doc`)
- Notes: Translated README files are support material for upstream docs and are not runtime parity requirements.

## Docs Assets And Branding Images

- ID: `23-docs-assets`
- Upstream refs:
  - `references/Scrapling/docs/assets/cover_dark.png` (`asset`)
  - `references/Scrapling/docs/assets/cover_dark.svg` (`asset`)
  - `references/Scrapling/docs/assets/cover_light.png` (`asset`)
  - `references/Scrapling/docs/assets/cover_light.svg` (`asset`)
  - `references/Scrapling/docs/assets/favicon.ico` (`asset`)
  - `references/Scrapling/docs/assets/logo.png` (`asset`)
  - `references/Scrapling/docs/assets/main_cover.png` (`asset`)
  - `references/Scrapling/docs/assets/scrapling_shell_curl.png` (`asset`)
  - `references/Scrapling/docs/assets/spider_architecture.png` (`asset`)
- Notes: Docs images and icons are visual support material, not Go parity behavior.
