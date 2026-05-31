# Project Boundaries

goscrapling is a Go-native Scrapling-style feature port. Scrapling is the
parity oracle, not the package architecture template.

## Product Positioning Boundary

The project is worth building as an evidence-backed Go-native extraction
engine for portfolio and Gormes/OpenClaw use. It should not be presented as a
complete Scrapling clone until parser, adaptive storage, fetchers,
browser-backed fetching, spiders, CLI surfaces, and tool integrations have
passing tests and completed progress rows.

Preferred external positioning:

> goscrapling is a Go-native web extraction engine inspired by Scrapling, built
> for agent runtimes and single-binary deployments.

Do not claim stealth, Cloudflare-solving, production proxy rotation, or complete
Scrapling parity until those rows are explicitly implemented, tested, and
documented.

## Scrapling Compatibility Boundary

`compat/scrapling` is allowed for upstream evidence, drift manifests, fixtures,
and explicit compatibility shims. It is not the destination for durable runtime
code.

Long-term Go architecture should use package names that describe the product
responsibility while remaining recognizable against Scrapling's taxonomy:

- `parser` for `Document`, `Element`, `Selection`, CSS/XPath, traversal,
  selector extraction, and adaptive relocation methods.
- `core/customtypes` for `TextHandler`, `TextHandlers`, and
  `AttributesHandler`.
- `core/storage` for fingerprints, scoring, adaptive store interfaces, and
  durable store adapters.
- `engines/toolbelt` for response construction and response metadata helpers.
- `engines/browser` for dynamic browser-backed fetching behind engine
  interfaces and concrete browser adapters.
- `fetchers` for static HTTP fetching, request options, sessions, proxy
  options, and fetcher errors.
- `spiders` for crawling requests, scheduler, session routing, results, robots,
  cache, and checkpoints.
- `cmd/goscrapling` for CLI surfaces.

The root package should act as a compatibility facade while the implementation
moves into taxonomy packages. Keep root re-exports until a future progress row
explicitly approves a breaking public API migration. Do not split packages just
to mirror upstream Python private files.

## Upstream Source Boundary

`references/Scrapling` is ignored study material. It is not vendored source.
Docs may cite upstream files and public behavior, but Go code should not copy
Python implementation details.

Port contracts, not file shape:

- preserve public parser, response, fetcher, browser, and spider behavior;
- translate Python class-level configuration into explicit Go options and
  sessions;
- translate async Python spider callbacks into Go interfaces or callback
  functions with context-aware execution;
- keep tests as the evidence for every parity claim.

## Package Naming Boundary

Donor names can appear where they are user-facing Scrapling concepts:
`Fetcher`, `Response`, `Spider`, `Request`, `Scheduler`, and adaptive
selection terms are acceptable exported names.

Donor-internal names should not leak into exported Go APIs unless a behavior
atom explains why the name is part of the user-facing contract.

## Safety Boundary

Stealth, proxy rotation, browser automation, and crawling controls require
explicit rows, tests, and operator-visible configuration. They should not be
added as incidental helpers while implementing static fetching or response
behavior.

Browser context/resource controls are explicit `BrowserOptions`, not hidden
stealth defaults. User agent, cookies, locale, timezone, proxy, CDP URL, real
Chrome selection, DNS-over-HTTPS, blocked domains, and ad/resource blocking
must be visible in `BrowserRequest` and fixture-backed before stronger browser
claims are made. The Go-owned ad-block boundary is intentionally small and must
not copy upstream's full list without a licensing decision.

Static identity controls are explicit request options. `StealthyHeaders` only
adds deterministic browser-like HTTP headers and a Google referer when callers
opt in; it does not claim TLS/browser fingerprint impersonation. Static
`Impersonate` and `HTTP3` options fail with operator-visible request-option
errors until a Go dependency and tests can prove those behaviors honestly.

Browser stealth controls are explicit `BrowserStealthOptions`, not defaults.
They may generate deterministic browser-like headers, opt into a Google
referer, pass WebRTC/WebGL launch controls, and install a canvas-noise init
script, but they do not claim automatic challenge solving. `SolveCloudflare`
returns an operator-visible unsupported browser option error before browser
engine work. `CloudflareChallengeBoundary()` is the inspectable status surface:
unsupported and disabled by default until a future row adds deterministic local
challenge fixtures, explicit tests, and operator-visible controls before any
Cloudflare/Turnstile bypass claim ships.

Static fetch safety controls are explicit request options. Robots enforcement
and private-network/CIDR blocking are opt-in, return typed `FetcherError`
values, and must be proven with `httptest` or pre-dispatch fixtures instead of
live network probes. Browser integration for the same safety policies requires
a separate browser-scoped row.

Core tests must stay hermetic:

- parser behavior uses string fixtures;
- fetcher behavior uses `httptest`;
- storage behavior uses `t.TempDir`;
- browser behavior uses fake browser engines before real engine smoke tests;
- spider behavior uses fake fetchers and deterministic scheduler fixtures.

## Integration Boundary

Gormes/OpenClaw integration must depend on goscrapling library APIs. The core
library must not import Gormes runtime packages. Gormes currently integrates
goscrapling behind its existing `web_extract` tool for static extraction and
selector evidence; any future standalone adapter package belongs outside the
core library.

Gormes owns tool registration, approval policy, truncation, channel rendering,
and typed unavailable results. goscrapling owns extraction behavior, response
construction, selector APIs, browser fetcher contracts, and spider primitives.
The first integration slice proves static fetch plus CSS extraction from local
fixtures before browser or crawler behavior is exposed to Gormes.
