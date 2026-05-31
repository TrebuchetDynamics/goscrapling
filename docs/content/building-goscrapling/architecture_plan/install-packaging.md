# Install And Packaging Boundary

This page records the Go-owned install and packaging behavior for the
Scrapling CLI parity row.

## Upstream Behavior

Scrapling exposes `scrapling install` as a utility command. At the observed
baseline it installs Playwright Chromium, Playwright system dependencies, and
TLD data, and the upstream Dockerfile bakes those dependencies into the image.
That behavior is appropriate for Scrapling's Python/Playwright stack.

## goscrapling Behavior

`goscrapling install` is intentionally non-mutating. It does not run package
managers, download browsers, install Docker assets, or inspect live host state.
It prints deterministic guidance for:

- installing the CLI with `go install`;
- static extraction and library APIs that need only Go module dependencies;
- browser extract modes that require Chrome or Chromium at runtime when used;
- container images that should copy the goscrapling binary and include
  Chrome/Chromium only when browser modes are needed.

Use `goscrapling install --json` for machine-readable dependency metadata. The
JSON marks `browser_downloads` as `false` so callers do not mistake the command
for an installer.

## Docker Guidance

This repository does not publish an official goscrapling Docker image yet. A
future packaging row may add one. Until then, downstream images should be
explicit about their runtime surface:

- static-only images can contain just the goscrapling binary and CA roots;
- browser-enabled images must include a Chrome/Chromium runtime compatible with
  chromedp;
- no image should imply Cloudflare solving, stealth bypass, proxy rotation, or
  live-service access unless a future tested row adds those controls.

This preserves Scrapling-visible CLI discoverability while avoiding hidden
browser installation side effects during normal Go binary use.
