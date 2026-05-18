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
