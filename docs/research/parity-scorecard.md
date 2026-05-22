# goscrapling Parity Scorecard

Generated from `progress.json` and benchmark fixture names under `benchmarks/`.
This scorecard is evidence for current Go port coverage; it is not a complete Scrapling parity claim.

## Upstream Benchmark Anchors

Source: `references/Scrapling/docs/benchmarks.md`.

- Text Extraction Speed Test: upstream reports Scrapling at 2.02 ms for 5000 nested elements.
- Element Similarity & Text Search Performance: upstream reports Scrapling at 2.39 ms.
- Local Go timings should be captured separately with `go test ./benchmarks -bench . -benchmem`.

## Coverage Scorecard

| Area | Status | Complete | In progress | Planned | Benchmark fixture |
|---|---:|---:|---:|---:|---|
| Parser and selectors | partial | 7 | 0 | 1 | `BenchmarkParserNestedText` |
| Static fetcher and response | partial | 9 | 0 | 3 | `BenchmarkStaticFetcherLocalResponse` |
| Spider runtime | partial | 4 | 0 | 6 | `BenchmarkSpiderSchedulerFingerprint` |
| CLI shell and extract commands | partial | 4 | 0 | 2 | `BenchmarkCLIExtractFixture` |

## Benchmark Fixtures

- `BenchmarkCLIExtractFixture`
- `BenchmarkParserNestedText`
- `BenchmarkSpiderSchedulerFingerprint`
- `BenchmarkStaticFetcherLocalResponse`

No live network, live browser, or live LLM is required: fixtures use in-memory HTML, `httptest`, deterministic spider scheduler data, and local CLI output files.
