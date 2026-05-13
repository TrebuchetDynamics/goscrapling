# goscrapling

Go-native adaptive web scraping inspired by D4Vinci/Scrapling.

`goscrapling` is planned as a Go-native adaptive web scraping framework focused on resilient element selection, fingerprint-based relocation, browser-backed fetching, and production-friendly crawling.

This project is not affiliated with D4Vinci/Scrapling. It is a study-driven Go implementation of Scrapling-style ideas, not an official port.

## Current Status

Planning and documentation phase.

The first implementation target is the adaptive parser MVP:

- Parse static HTML into queryable selectors.
- Select elements with CSS-like APIs.
- Save an element fingerprint under a domain plus identifier.
- Relocate the same logical element after markup changes.
- Drive all behavior with test-first fixtures.

## Reference Material

The upstream Scrapling repository is cloned locally for study at:

`references/Scrapling`

The local clone is ignored by git. Public documentation records only the observed architecture and decisions, not copied source.

## Documentation

- [Scrapling Architecture Map](docs/research/scrapling-architecture-map.md)
- [Go Scraping OSS Survey](docs/research/go-scraping-oss-survey.md)
- [Adaptive Parser MVP Design](docs/superpowers/specs/2026-05-13-goscrapling-adaptive-parser-design.md)
