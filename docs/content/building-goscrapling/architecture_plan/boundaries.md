# Project Boundaries

goscrapling is a Go-native Scrapling-style feature port. Scrapling is the
parity oracle, not the package architecture template.

## Scrapling Compatibility Boundary

`compat/scrapling` is allowed for upstream evidence, drift manifests, fixtures,
and explicit compatibility shims. It is not the destination for durable runtime
code.

Long-term Go architecture should use package names that describe the product
responsibility:

- `adaptive` for fingerprints, relocation, and storage adapters when the root
  package becomes too crowded.
- `fetcher` for static HTTP fetching, response construction, request options,
  sessions, and fetcher errors.
- `browser` for dynamic browser-backed fetching behind engine interfaces.
- `spider` for crawling requests, scheduler, session routing, results, robots,
  cache, and checkpoints.
- `storage` for durable stores if adaptive storage outgrows the root package.
- `cmd/goscrapling` for CLI surfaces.

The root package may continue to host the current parser and adaptive API until
there is enough implementation pressure to split it. Do not split packages just
to mirror upstream Python files.

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
library must not import Gormes runtime packages. A future tool adapter belongs
under `integrations/gormes` or a separate command surface after fetcher and
response behavior are stable.
