# goscrapling Upstream App Map Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a validated full-app upstream Scrapling map, with py2many static reference outputs kept out of production Go packages.

**Architecture:** Extend the existing `internal/progress` and `cmd/progress` control plane. The app map lives as JSON source of truth under `docs/content/building-goscrapling/architecture_plan/`, renders to Markdown for humans, and validates against the local `references/Scrapling` inventory when that checkout exists.

**Tech Stack:** Go standard library, existing `internal/progress` package, `cmd/progress`, `jq`, optional `py2many --go` CLI used only for docs/research reference output.

---

## File Structure

- Create `internal/progress/appmap.go`: app-map model, JSON loading, validation, upstream inventory discovery, coverage validation, Markdown rendering.
- Create `internal/progress/appmap_test.go`: app-map unit tests for validation, coverage, and rendering.
- Modify `cmd/progress/main.go`: add `map-validate` and `map-write` commands.
- Modify `cmd/progress/main_test.go`: add CLI tests for app-map validation and Markdown generation.
- Create `docs/content/building-goscrapling/architecture_plan/upstream-app-map.json`: canonical full-app map.
- Create `docs/content/building-goscrapling/architecture_plan/upstream-app-map.md`: generated human view.
- Create `docs/research/python-to-go-probes/py2many/summary.md`: py2many reference summary.
- Create `docs/research/python-to-go-probes/py2many/go/*.go.txt`: py2many stdout captures with reference-only headers where conversion succeeds or partially emits output.
- Modify `docs/content/building-goscrapling/architecture_plan/progress.json`: add a complete docs/control-plane row for the upstream app map.
- Regenerate `docs/content/building-goscrapling/builder-loop/*.md` with `go run ./cmd/progress write`.

## Task 1: Capture py2many Static Reference Outputs

**Files:**
- Create: `docs/research/python-to-go-probes/py2many/summary.md`
- Create: `docs/research/python-to-go-probes/py2many/go/*.go.txt`

- [ ] **Step 1: Create the output directory**

Run:

```sh
mkdir -p docs/research/python-to-go-probes/py2many/go
```

Expected: directory exists and is outside all production Go packages.

- [ ] **Step 2: Check py2many availability**

Run:

```sh
if command -v py2many >/dev/null 2>&1; then
  py2many --help | sed -n '1,40p'
else
  python3 -m pip --version
fi
```

Expected: either local `py2many` help prints, or Python/pip availability is recorded for fallback installation.

- [ ] **Step 3: Run py2many as a static reference probe**

Use py2many's documented Go backend command shape, `py2many --go <file>`, from the upstream README.

Run this loop:

```sh
set -u
inputs='
references/Scrapling/scrapling/parser.py
references/Scrapling/scrapling/core/custom_types.py
references/Scrapling/scrapling/core/shell.py
references/Scrapling/scrapling/core/ai.py
references/Scrapling/scrapling/cli.py
references/Scrapling/scrapling/engines/static.py
references/Scrapling/scrapling/engines/toolbelt/proxy_rotation.py
references/Scrapling/scrapling/engines/toolbelt/navigation.py
references/Scrapling/scrapling/engines/_browsers/_validators.py
references/Scrapling/scrapling/spiders/robotstxt.py
references/Scrapling/scrapling/spiders/cache.py
references/Scrapling/scrapling/spiders/checkpoint.py
references/Scrapling/scrapling/spiders/links.py
references/Scrapling/scrapling/spiders/templates/crawler.py
references/Scrapling/scrapling/spiders/templates/sitemap.py
'
for input in $inputs; do
  out="docs/research/python-to-go-probes/py2many/go/$(echo "$input" | sed 's#^references/Scrapling/scrapling/##; s#/#__#g; s#\.py$#.go.txt#')"
  {
    printf '// Reference-only py2many output. Do not copy into production packages.\n'
    printf '// Source: %s\n\n' "$input"
    if command -v py2many >/dev/null 2>&1; then
      py2many --go "$input"
    else
      python3 -m pipx run py2many --go "$input"
    fi
  } >"$out" 2>"$out.stderr" || true
done
```

Expected: the loop exits successfully even when individual translations fail. Some `.go.txt` files may contain only the reference header if py2many cannot translate a file.

- [ ] **Step 4: Write the py2many summary**

Create `docs/research/python-to-go-probes/py2many/summary.md` with this shape:

```markdown
# py2many Static Reference Probe

Date: 2026-05-14

Purpose: reference-only Python-to-Go static translation evidence for the goscrapling upstream app map.

Command shape: `py2many --go <python-file>`

Source: https://github.com/py2many/py2many

Policy:

- Generated output is reference evidence only.
- Generated output must not be copied into production Go packages.
- Parity still requires progress rows and Go tests.

## Inputs

| Input | Output | Result |
|---|---|---|
```

Append one row per probed file. Mark the result as `output`, `stderr`, or `missing-tool` based on the captured files.

- [ ] **Step 5: Sanity check generated artifacts**

Run:

```sh
find docs/research/python-to-go-probes/py2many -type f | sort
rg -n "Reference-only py2many output|Generated output is reference evidence only" docs/research/python-to-go-probes/py2many
```

Expected: outputs are under `docs/research/python-to-go-probes/py2many/` and every generated Go capture is clearly marked reference-only.

## Task 2: Add App-Map Model, Validation, Inventory, And Rendering

**Files:**
- Create: `internal/progress/appmap_test.go`
- Create: `internal/progress/appmap.go`

- [ ] **Step 1: Write failing tests**

Create `internal/progress/appmap_test.go` with tests named:

```go
func TestValidateAppMapRejectsIncompleteEntries(t *testing.T)
func TestValidateAppMapCoverageFindsUnmappedUpstreamRefs(t *testing.T)
func TestRenderAppMapMarkdownIncludesEntries(t *testing.T)
```

The tests should exercise these APIs before they exist:

```go
err := ValidateAppMap(&AppMap{})
err := ValidateAppMapCoverage(root, fixtureAppMap())
body := RenderAppMapMarkdown(fixtureAppMap())
```

Use a temp `references/Scrapling` fixture with:

```text
references/Scrapling/scrapling/parser.py
references/Scrapling/tests/parser/test_general.py
references/Scrapling/docs/parsing/main_classes.md
```

The coverage test should first omit one file and expect an error containing `unmapped upstream ref`, then include all three and expect no error.

- [ ] **Step 2: Run focused test and confirm RED**

Run:

```sh
go test ./internal/progress -run 'TestValidateAppMap|TestRenderAppMap' -count=1
```

Expected: FAIL because `AppMap`, `ValidateAppMap`, `ValidateAppMapCoverage`, and `RenderAppMapMarkdown` do not exist.

- [ ] **Step 3: Implement app-map types and loading**

Create `internal/progress/appmap.go` with these public types and functions:

```go
type AppMap struct {
    Meta    AppMapMeta    `json:"meta"`
    Entries []AppMapEntry `json:"entries"`
}

type AppMapMeta struct {
    Version           string       `json:"version"`
    Upstream          UpstreamMeta `json:"upstream,omitempty"`
    GeneratedMarkdown string       `json:"generated_markdown,omitempty"`
    Py2ManyProbeDir   string       `json:"py2many_probe_dir,omitempty"`
}

type AppMapEntry struct {
    ID                     string      `json:"id"`
    Title                  string      `json:"title"`
    Upstream               []AppMapRef `json:"upstream"`
    FeatureAnchor          string      `json:"feature_anchor,omitempty"`
    BehaviorAtoms          []string    `json:"behavior_atoms,omitempty"`
    GoTarget               string      `json:"go_target,omitempty"`
    ProgressRows           []string    `json:"progress_rows,omitempty"`
    CoverageStatus         string      `json:"coverage_status"`
    TranslationSuitability string      `json:"translation_suitability"`
    StaticReferencePaths   []string    `json:"static_reference_paths,omitempty"`
    Notes                  []string    `json:"notes,omitempty"`
}

type AppMapRef struct {
    Ref     string   `json:"ref"`
    Kind    string   `json:"kind"`
    Symbols []string `json:"symbols,omitempty"`
}

func LoadAppMap(path string) (*AppMap, error)
func ValidateAppMap(m *AppMap) error
func ValidateAppMapCoverage(repoRoot string, m *AppMap) error
func RenderAppMapMarkdown(m *AppMap) string
```

- [ ] **Step 4: Implement validation rules**

Validation must reject:

- missing `meta.version`;
- empty `entries`;
- duplicate entry IDs;
- entries with no upstream refs;
- upstream refs without `ref` or `kind`;
- invalid ref kind outside `source`, `test`, `doc`, `asset`;
- invalid coverage status outside `covered`, `partial`, `planned`, `vague`, `owned`, `excluded`;
- invalid translation suitability outside `probe_candidate`, `manual_rewrite`, `not_useful`, `not_applicable`;
- non-excluded entries missing `feature_anchor`, `go_target`, or `progress_rows`.

- [ ] **Step 5: Implement inventory coverage**

`ValidateAppMapCoverage(repoRoot, m)` should return nil when `references/Scrapling` does not exist. When it exists, it should inventory:

```text
references/Scrapling/scrapling/**/*.py
references/Scrapling/scrapling/py.typed
references/Scrapling/tests/**/*.py
references/Scrapling/docs/**/*.md
references/Scrapling/docs/assets/**/*.{png,svg,ico}
```

Each discovered path must appear in at least one `AppMapRef.Ref`. Missing refs should be reported as `unmapped upstream ref <path>`.

- [ ] **Step 6: Implement Markdown rendering**

`RenderAppMapMarkdown` should render:

- title and source metadata;
- a summary count by coverage status;
- a table with entry title, status, feature anchor, Go target, progress rows, translation suitability, and upstream ref count;
- one detail section per entry listing upstream refs and static reference paths.

- [ ] **Step 7: Run focused test and confirm GREEN**

Run:

```sh
go test ./internal/progress -run 'TestValidateAppMap|TestRenderAppMap' -count=1
```

Expected: PASS.

## Task 3: Add `cmd/progress map-validate` And `map-write`

**Files:**
- Modify: `cmd/progress/main_test.go`
- Modify: `cmd/progress/main.go`

- [ ] **Step 1: Write failing CLI tests**

Add tests named:

```go
func TestRunMapValidate(t *testing.T)
func TestRunMapWriteRegeneratesMarkdown(t *testing.T)
```

`TestRunMapValidate` should call:

```go
err := run(&stdout, &stderr, []string{"--repo-root", "../..", "map-validate"})
```

and expect stdout to contain `app-map: validated`.

`TestRunMapWriteRegeneratesMarkdown` should create a temp repo root, copy `upstream-app-map.json` into it, run:

```go
err := run(&stdout, &stderr, []string{"--repo-root", root, "map-write"})
```

and assert `docs/content/building-goscrapling/architecture_plan/upstream-app-map.md` exists and contains `# Upstream Scrapling App Map`.

- [ ] **Step 2: Run focused test and confirm RED**

Run:

```sh
go test ./cmd/progress -run 'TestRunMap' -count=1
```

Expected: FAIL because the commands are not implemented or the manifest does not exist yet.

- [ ] **Step 3: Implement command dispatch**

In `cmd/progress/main.go`, extend `usage` to:

```go
const usage = "usage: progress [--repo-root <path>] {validate|write|map-validate|map-write}"
```

Add cases:

```go
case "map-validate":
    appMap, err := loadValidAppMap(root)
    if err != nil {
        return err
    }
    _, err = fmt.Fprintf(stdout, "app-map: validated %d entries\n", len(appMap.Entries))
    return err
case "map-write":
    return writeAppMap(stdout, root)
```

Add helpers:

```go
func loadValidAppMap(root string) (*progress.AppMap, error)
func writeAppMap(stdout io.Writer, root string) error
```

`writeAppMap` should write `RenderAppMapMarkdown(appMap)` directly to:

```text
docs/content/building-goscrapling/architecture_plan/upstream-app-map.md
```

- [ ] **Step 4: Run focused test and confirm GREEN after Task 4 manifest exists**

Run:

```sh
go test ./cmd/progress -run 'TestRunMap' -count=1
```

Expected after Task 4 creates the manifest: PASS.

## Task 4: Create The Full Upstream App Map Manifest

**Files:**
- Create: `docs/content/building-goscrapling/architecture_plan/upstream-app-map.json`
- Create: `docs/content/building-goscrapling/architecture_plan/upstream-app-map.md`
- Modify: `docs/content/building-goscrapling/architecture_plan/progress.json`
- Modify generated: `docs/content/building-goscrapling/builder-loop/*.md`

- [ ] **Step 1: Inventory upstream refs**

Run:

```sh
find references/Scrapling/scrapling -type f \( -name '*.py' -o -name 'py.typed' \) | sort
find references/Scrapling/tests -type f -name '*.py' | sort
find references/Scrapling/docs -type f \( -name '*.md' -o -path '*/assets/*.png' -o -path '*/assets/*.svg' -o -path '*/assets/*.ico' \) | sort
```

Expected: inventory lists the upstream source, tests, docs, and docs assets that must be covered by the manifest.

- [ ] **Step 2: Create grouped manifest entries**

Create `upstream-app-map.json` with entries for these app areas:

- package metadata and public exports;
- parser, selectors, custom types, translator, and adaptive storage;
- static response and fetchers;
- browser and stealth fetchers;
- spider core and crawler engine;
- robots/cache/checkpoint/production crawler controls;
- link extraction and spider templates;
- CLI extract and shell;
- MCP/AI tools;
- upstream tests grouped by subsystem;
- upstream docs grouped by subsystem;
- translated README files and docs assets as excluded support material.

Each discovered upstream file from Step 1 must appear in exactly one or more `upstream[].ref` values.

- [ ] **Step 3: Link py2many reference outputs**

For entries whose source files were probed in Task 1, add `static_reference_paths` pointing at the relevant `.go.txt` or `summary.md` files.

Mark `translation_suitability` conservatively:

- `probe_candidate` for small, typed-ish helper/control files;
- `manual_rewrite` for dynamic parser, browser, async, CLI, shell, MCP, and spider engine surfaces;
- `not_useful` for huge constants or Python-runtime-specific behavior;
- `not_applicable` for tests, docs, assets, and excluded rows.

- [ ] **Step 4: Add a completed progress row**

Add a `docs`-owned complete row under `phase-5-cli-tooling` / `tool-surfaces`:

```json
{
  "name": "Full upstream app map and py2many static reference",
  "status": "complete",
  "priority": "P1",
  "contract": "Validate a full upstream Scrapling app map covering source, tests, docs, and reference-only py2many output before future parity planning.",
  "contract_status": "validated",
  "slice_size": "medium",
  "execution_owner": "docs",
  "source_refs": [
    "references/Scrapling",
    "docs/superpowers/specs/2026-05-14-goscrapling-upstream-app-map-design.md",
    "https://github.com/py2many/py2many"
  ],
  "ready_when": [
    "The upstream Scrapling checkout exists at references/Scrapling."
  ],
  "write_scope": [
    "internal/progress/appmap.go",
    "internal/progress/appmap_test.go",
    "cmd/progress/main.go",
    "cmd/progress/main_test.go",
    "docs/content/building-goscrapling/architecture_plan/upstream-app-map.json",
    "docs/content/building-goscrapling/architecture_plan/upstream-app-map.md",
    "docs/research/python-to-go-probes/py2many/"
  ],
  "test_commands": [
    "go test ./... -run 'TestValidateAppMap|TestRenderAppMap|TestRunMap' -count=1",
    "go run ./cmd/progress map-validate"
  ],
  "acceptance": [
    "Every upstream source, test, docs, and docs asset ref is covered by the manifest or explicitly excluded.",
    "py2many outputs are stored only as reference evidence outside production Go packages.",
    "map-validate and map-write are available through cmd/progress."
  ],
  "done_signal": [
    "Focused app-map tests pass, map validation passes, and generated Markdown is regenerated from the JSON manifest."
  ]
}
```

- [ ] **Step 5: Generate Markdown and builder-loop docs**

Run:

```sh
go run ./cmd/progress map-validate
go run ./cmd/progress map-write
go run ./cmd/progress write
```

Expected: app-map validation passes, `upstream-app-map.md` is generated, and builder-loop docs reflect the new complete row.

## Task 5: Final Validation And Cleanup

**Files:**
- Verify all files changed by Tasks 1-4.

- [ ] **Step 1: Run focused tests**

Run:

```sh
go test ./internal/progress -run 'TestValidateAppMap|TestRenderAppMap' -count=1
go test ./cmd/progress -run 'TestRunMap' -count=1
```

Expected: both commands pass.

- [ ] **Step 2: Run repo validation commands**

Run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress map-validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
jq empty docs/content/building-goscrapling/architecture_plan/upstream-app-map.json
git diff --check
```

Expected: all commands exit 0.

- [ ] **Step 3: Inspect changed files**

Run:

```sh
git status --short
git diff --stat
git diff -- docs/content/building-goscrapling/architecture_plan/upstream-app-map.json | sed -n '1,220p'
git diff -- internal/progress/appmap.go cmd/progress/main.go | sed -n '1,260p'
```

Expected: changed files are limited to the app-map control plane, app-map docs, py2many reference outputs, progress row, and generated builder-loop docs. Existing unrelated dirty files remain untouched.

## Plan Self-Review

- Spec coverage: Tasks 1-4 cover py2many static evidence, JSON source of truth, generated Markdown, validation, command integration, and progress-row ownership. Task 5 covers required verification.
- Red-flag scan: no unresolved planning language remains. The manifest grouping is explicit enough for implementation while allowing exact refs to be copied from the inventory command.
- Type consistency: App-map type and command names are consistent across tests, implementation, and validation commands.
