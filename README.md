# goscrapling

Go-native adaptive web scraping inspired by D4Vinci/Scrapling.

`goscrapling` is a Go-native adaptive web scraping framework focused on resilient element selection. The current MVP parses static HTML, saves element fingerprints, and relocates elements after markup changes.

This project is not affiliated with D4Vinci/Scrapling. It is a study-driven Go implementation of Scrapling-style ideas, not an official port.

## Current Status

Adaptive parser MVP implemented.

Implemented now:

- Parse static HTML into queryable selectors.
- Select elements with CSS-like APIs.
- Save an element fingerprint under a domain plus identifier.
- Relocate the same logical element after markup changes.
- Drive all behavior with test-first fixtures.

Not implemented yet:

- HTTP fetchers.
- Browser-backed fetching.
- Crawling and spider scheduling.
- CLI, MCP, or Gormes/OpenClaw tool integration.

## Example

```go
ctx := context.Background()
store := goscrapling.NewMemoryStore()

before, _ := goscrapling.Parse(strings.NewReader(`<article class="product" id="p1">Product 1</article>`), goscrapling.ParseOptions{
    URL:   "https://example.com/products",
    Store: store,
})
element, _ := before.CSS("#p1").First()
_ = before.Save(ctx, element, "featured-product")

after, _ := goscrapling.Parse(strings.NewReader(`<article class="product" data-id="p1"><span>Product 1</span></article>`), goscrapling.ParseOptions{
    URL:   "https://example.com/products",
    Store: store,
})
match, ok, _ := after.Relocate(ctx, "featured-product")
fmt.Println(ok, match.Element.Text())
```

## Reference Material

The upstream Scrapling repository is cloned locally for study at:

`references/Scrapling`

The local clone is ignored by git. Public documentation records only the observed architecture and decisions, not copied source.

## Documentation

- [Scrapling Architecture Map](docs/research/scrapling-architecture-map.md)
- [Go Scraping OSS Survey](docs/research/go-scraping-oss-survey.md)
- [Adaptive Parser MVP Design](docs/superpowers/specs/2026-05-13-goscrapling-adaptive-parser-design.md)
