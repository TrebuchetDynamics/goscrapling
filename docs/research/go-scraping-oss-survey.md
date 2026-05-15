# Go Scraping OSS Survey

Date: 2026-05-13

Purpose: identify useful open-source Go projects before designing `goscrapling`.

## Primary References

### gocolly/colly

- URL: `https://github.com/gocolly/colly`
- Role: mature crawler and scraper framework.
- Useful lessons: request lifecycle, callbacks, collector ergonomics, caching, storage, rate limiting.
- Risk: higher-level crawler model can pull the MVP away from adaptive element relocation.

### PuerkitoBio/goquery

- URL: `https://github.com/PuerkitoBio/goquery`
- Role: jQuery-like DOM query library for Go.
- Useful lessons: selection API ergonomics and HTML traversal.
- Risk: does not solve adaptive matching by itself.

### chromedp/chromedp

- URL: `https://github.com/chromedp/chromedp`
- Role: Chrome DevTools Protocol automation.
- Useful lessons: browser-backed fetching, screenshot/debug tooling, CDP control.
- Risk: lower-level API may require more wrapper code for scraper ergonomics.

### go-rod/rod

- URL: `https://github.com/go-rod/rod`
- Role: browser automation and scraping through CDP.
- Useful lessons: high-level page operations, browser lifecycle, scraping examples.
- Risk: browser sessions, context/resource controls, and XHR capture still need focused slices on top of the selected `chromedp` backend.

### MontFerret/ferret

- URL: `https://github.com/MontFerret/ferret`
- Role: declarative web scraping.
- Useful lessons: query-oriented extraction language and execution model.
- Risk: declarative language is out of scope for the first MVP.

### geziyor/geziyor

- URL: `https://github.com/geziyor/geziyor`
- Role: Go crawling framework with JavaScript rendering support.
- Useful lessons: spider/crawler organization and JS rendering integration.
- Risk: crawler-first shape is broader than the first adaptive parser target.

### enetx/surf

- URL: `https://github.com/enetx/surf`
- Role: advanced HTTP client with browser impersonation and HTTP/3 support.
- Useful lessons: future static fetcher stealth and TLS fingerprinting.
- Risk: anti-bot claims and dependency tradeoffs need separate validation.

### projectdiscovery/katana

- URL: `https://github.com/projectdiscovery/katana`
- Role: crawling and spidering framework.
- Useful lessons: production crawling, URL discovery, performance, CLI patterns.
- Risk: security-scanner crawler priorities differ from adaptive extraction.

## Secondary References

### gosom/scrapemate

- URL: `https://github.com/gosom/scrapemate`
- Role: Go crawling and scraping framework.
- Useful lessons: small framework design and crawling abstractions.

### JSLEEKR/scrapling-go

- URL: `https://github.com/JSLEEKR/scrapling-go`
- Role: existing Go reimplementation name-collision reference.
- Useful lessons: avoid exact name collision and study gaps.
- Risk: do not depend on or represent it as upstream.

### sadewadee/foxhound

- URL: `https://github.com/sadewadee/foxhound`
- Role: newer Go scraping framework claiming adaptive parsing and anti-detection.
- Useful lessons: compare API shape and test strategy before browser/stealth phases.
- Risk: new, low-star project; claims require source review and tests.

## First Dependency Hypothesis

For the adaptive parser MVP:

- Start with Go standard library plus a mature HTML parser.
- Use `goquery` only if it gives enough selector ergonomics without hiding the DOM details needed for fingerprinting.
- Keep browser and crawler dependencies out of phase 1.

For later phases:

- Deepen the selected `chromedp` browser backend with sessions, context/resource controls, and XHR capture.
- Compare `colly`, `geziyor`, and `katana` before building spider scheduling.
- Evaluate `surf` only when static fetch stealth becomes a concrete requirement.

## Selection Criteria

Each dependency must be judged by:

- License compatibility.
- Maintenance activity.
- API stability.
- Testability without live network access.
- Ability to expose low-level DOM or response metadata.
- Fit with Go-native interfaces and `context.Context`.
