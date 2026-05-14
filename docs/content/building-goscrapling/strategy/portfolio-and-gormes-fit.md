# Portfolio and Gormes Fit

goscrapling is worth building if it stays positioned as an evidence-backed
Go-native web extraction engine, not as a quick complete clone of Scrapling.

The project is portfolio-worthy because it demonstrates a hard engineering
loop: study a mature upstream surface, split parity into tested slices, keep a
progress ledger, and ship a Go implementation that can be embedded in larger
agent systems. It stops being portfolio-worthy if it becomes a broad scraper
with unsupported claims, vague parity language, or untested anti-bot features.

## Product Thesis

Recommended positioning:

> goscrapling is a Go-native web extraction engine inspired by Scrapling, built
> for agent runtimes and single-binary deployments.

Use this language when presenting the project:

- "Go-native Scrapling-style extraction engine"
- "evidence-backed port toward selected Scrapling behavior"
- "agent-ready parser, fetcher, response, browser, and spider substrate"
- "single-binary-friendly scraping core for Gormes/OpenClaw-style runtimes"

Avoid this language until the progress ledger and tests prove it:

- "complete Scrapling clone"
- "full Scrapling port"
- "Cloudflare bypass engine"
- "production stealth scraper"
- "drop-in replacement for Scrapling"

## Portfolio Value

The portfolio story is strong because the work is not just another scraper.
The visible engineering artifacts matter:

- a public upstream feature map instead of hand-wavy scope;
- a progress ledger that turns a large port into builder-sized rows;
- hermetic parser, fetcher, storage, browser-contract, CLI, and spider tests;
- generated agent queue docs for parallel implementation;
- explicit safety boundaries for browser, proxy, stealth, and crawler behavior;
- a concrete downstream consumer: Gormes.

The portfolio claim should be: "I can take a complex Python library and build
a Go-native, test-led port with traceable parity decisions." That is stronger
than claiming a finished scraper before the hard surfaces are complete.

## Gormes Fit

goscrapling fits Gormes because Gormes is a Go single-binary agent runtime and
needs reliable web extraction without pulling a Python runtime into the tool
loop. The integration value is:

- static fetch and CSS extraction for deterministic agent tools;
- structured evidence output with URL, status, selected text, and metadata;
- parser and response behavior usable inside provider/tool calls;
- future browser-backed extraction that can align with Gormes browser runtime
  checks;
- future crawling, cache, robots, and checkpoint behavior for research tasks;
- no loopback service or sidecar required for the first integration slice.

Gormes should own tool registration, approval policy, truncation, channel
rendering, and unavailable-result behavior. goscrapling should own extraction
behavior, response construction, selector APIs, browser fetcher contracts, and
spider/crawl primitives.

The core goscrapling package must not import Gormes runtime packages. A Gormes
adapter belongs under `integrations/gormes` or a separate command/package
surface after static fetcher and response behavior are stable.

## Recommended Milestones

1. Close the P0 parser and static fetcher gaps already listed in
   `progress.json`.
2. Build a fixture-backed `integrations/gormes` static extraction adapter that
   can fetch a local page, apply CSS selection, and return structured evidence.
3. Add output-shaping tests that prove the adapter is useful to an agent tool
   caller without requiring live web access.
4. Add browser-backed extraction only after the browser engine seam is stable
   and local dependency gating is documented.
5. Add crawler/cache/robots/checkpoint behavior after the spider rows move from
   planned to tested.

## Risk Boundaries

The main risk is scope. Scrapling includes parsing, adaptive relocation,
fetchers, browser automation, stealth controls, spiders, CLI, MCP, and Docker
surfaces. A portfolio/Gormes version should advance through narrow rows rather
than trying to market all of that at once.

Do not ship or document stealth, proxy rotation, Cloudflare-solving, or browser
automation behavior as available until the relevant row has explicit operator
controls, tests, docs, and a done signal. For now, those are parity targets and
planning rows.

## Evidence to Keep Current

Keep these surfaces synchronized when the strategy changes:

- `README.md`
- `docs/content/building-goscrapling/architecture_plan/boundaries.md`
- `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- `docs/content/building-goscrapling/architecture_plan/progress.json`
- `../gormes-agent/docs/content/architecture/tool-execution.md`
- `../gormes-agent/docs/content/building-gormes/cross-project-feature-map.md`
