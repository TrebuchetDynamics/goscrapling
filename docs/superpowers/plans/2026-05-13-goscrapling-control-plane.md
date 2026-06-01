# goscrapling Control Plane Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a small Gormes-style progress validator and generated builder queue so goscrapling can keep a true Scrapling port backlog without side queues.

**Architecture:** Add `internal/progress` for the progress model, validation, queue selection, markdown rendering, and marker replacement. Add `cmd/progress` as the operator entrypoint. Keep all generated docs under `docs/content/building-goscrapling/builder-loop/`, with `progress.json` as the only source of truth.

**Tech Stack:** Go standard library, `encoding/json`, `errors.Join`, `flag`-style argument parsing, existing Go test runner, existing markdown docs.

---

## File Structure

- Create `internal/progress/progress.go`: typed progress model, JSON loading, derived counts, row helpers.
- Create `internal/progress/validate.go`: schema validation for contract rows and metadata.
- Create `internal/progress/render.go`: markdown renderers for builder-loop handoff, queue, next slices, blocked rows, umbrella cleanup.
- Create `internal/progress/markers.go`: `ReplaceMarker` helper for `<!-- PROGRESS:START kind=... -->` sections.
- Create `internal/progress/progress_test.go`: tests for validation, queue selection, umbrella detection, and marker replacement.
- Create `cmd/progress/main.go`: `validate` and `write` CLI commands.
- Create `cmd/progress/main_test.go`: CLI tests using temp repos and markered docs.
- Modify `docs/content/building-goscrapling/architecture_plan/progress.json`: add `agent_queue`, split Phase 1 rows, keep future rows row-backed.
- Create generated docs:
  - `docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md`
  - `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`
  - `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`
  - `docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked-slices.md`
  - `docs/content/building-goscrapling/builder-loop/surfaces/cleanup/umbrella-cleanup.md`
- Create `docs/content/building-goscrapling/architecture_plan/scrapling-behavior-atoms.md`.
- Create `docs/content/building-goscrapling/architecture_plan/boundaries.md`.
- Modify `docs/content/building-goscrapling/builder-loop/schema/progress-schema.md`.
- Modify `progress_docs_test.go`: strengthen upstream coverage checks and call `internal/progress` validation.
- Modify `README.md` and `AGENTS.md`: mention `go run ./cmd/progress validate/write`.

## Task 1: Progress Package Tests First

**Files:**
- Create: `internal/progress/progress_test.go`
- Create later: `internal/progress/progress.go`
- Create later: `internal/progress/validate.go`
- Create later: `internal/progress/render.go`
- Create later: `internal/progress/markers.go`

- [ ] **Step 1: Write failing tests**

Create `internal/progress/progress_test.go` with tests named:

```go
func TestValidateAcceptsCurrentProgress(t *testing.T)
func TestValidateRejectsIncompleteContractRows(t *testing.T)
func TestAgentQueueExcludesCompleteBlockedAndUmbrellaRows(t *testing.T)
func TestRenderUmbrellaCleanupIncludesUmbrellaRows(t *testing.T)
func TestReplaceMarkerReplacesMatchingSection(t *testing.T)
```

The tests should call APIs that do not exist yet:

```go
progress, err := Load("../../docs/content/building-goscrapling/architecture_plan/progress.json")
if err != nil {
    t.Fatalf("Load: %v", err)
}
if err := Validate(progress); err != nil {
    t.Fatalf("Validate: %v", err)
}
rows := AgentQueueRows(progress)
body := RenderUmbrellaCleanup(progress)
out, err := ReplaceMarker(input, "agent-queue", "generated")
```

- [ ] **Step 2: Run focused test and confirm RED**

Run:

```sh
go test ./internal/progress -count=1
```

Expected: fail because `internal/progress` APIs are undefined.

- [ ] **Step 3: Implement `internal/progress`**

Add typed structs matching the existing JSON:

```go
type Status string
const (
    StatusComplete Status = "complete"
    StatusInProgress Status = "in_progress"
    StatusPlanned Status = "planned"
)

type Progress struct {
    Meta Meta `json:"meta"`
    Phases map[string]Phase `json:"phases"`
}

type Item struct {
    Name string `json:"name"`
    Priority string `json:"priority,omitempty"`
    Status Status `json:"status"`
    Contract string `json:"contract,omitempty"`
    ContractStatus string `json:"contract_status,omitempty"`
    SliceSize string `json:"slice_size,omitempty"`
    ExecutionOwner string `json:"execution_owner,omitempty"`
    SourceRefs []string `json:"source_refs,omitempty"`
    ReadyWhen []string `json:"ready_when,omitempty"`
    NotReadyWhen []string `json:"not_ready_when,omitempty"`
    BlockedBy []string `json:"blocked_by,omitempty"`
    Unblocks []string `json:"unblocks,omitempty"`
    WriteScope []string `json:"write_scope,omitempty"`
    TestCommands []string `json:"test_commands,omitempty"`
    NoTestRequired string `json:"no_test_required,omitempty"`
    Acceptance []string `json:"acceptance,omitempty"`
    DoneSignal []string `json:"done_signal,omitempty"`
}
```

Implement:

```go
func Load(path string) (*Progress, error)
func Validate(p *Progress) error
func AgentQueueRows(p *Progress) []Row
func BlockedRows(p *Progress) []Row
func UmbrellaRows(p *Progress) []Row
func RenderBuilderLoopHandoff(p *Progress) string
func RenderAgentQueue(p *Progress) string
func RenderNextSlices(p *Progress) string
func RenderBlockedSlices(p *Progress) string
func RenderUmbrellaCleanup(p *Progress) string
func ReplaceMarker(input, kind, body string) (string, error)
```

- [ ] **Step 4: Run focused test and confirm GREEN**

Run:

```sh
go test ./internal/progress -count=1
```

Expected: pass.

- [ ] **Step 5: Commit**

```sh
git add internal/progress
git commit -m "test: add progress control model"
```

## Task 2: Progress CLI With Validate And Write

**Files:**
- Create: `cmd/progress/main_test.go`
- Create: `cmd/progress/main.go`

- [ ] **Step 1: Write failing CLI tests**

Create tests that call:

```go
err := run(&stdout, &stderr, []string{"--repo-root", root, "validate"})
err := run(&stdout, &stderr, []string{"--repo-root", root, "write"})
```

The validate test expects output containing:

```text
progress: validated
```

The write test creates temp markered docs and expects generated content to
replace the `agent-queue` marker.

- [ ] **Step 2: Run focused test and confirm RED**

Run:

```sh
go test ./cmd/progress -count=1
```

Expected: fail because `cmd/progress` does not exist.

- [ ] **Step 3: Implement CLI**

Implement:

```go
const usage = "usage: progress [--repo-root <path>] {validate|write}"

func run(stdout, stderr io.Writer, args []string) error
```

`validate` loads:

```text
docs/content/building-goscrapling/architecture_plan/progress.json
```

`write` rewrites these files:

```text
docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md
docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md
docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md
docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked-slices.md
docs/content/building-goscrapling/builder-loop/surfaces/cleanup/umbrella-cleanup.md
```

- [ ] **Step 4: Run focused test and confirm GREEN**

Run:

```sh
go test ./cmd/progress -count=1
```

Expected: pass.

- [ ] **Step 5: Commit**

```sh
git add cmd/progress
git commit -m "feat: add progress control cli"
```

## Task 3: Builder-Loop Docs And Phase 1 Row Split

**Files:**
- Modify: `docs/content/building-goscrapling/architecture_plan/progress.json`
- Create: `docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md`
- Create: `docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md`
- Create: `docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md`
- Create: `docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked-slices.md`
- Create: `docs/content/building-goscrapling/builder-loop/surfaces/cleanup/umbrella-cleanup.md`
- Modify: `docs/content/building-goscrapling/builder-loop/schema/progress-schema.md`

- [ ] **Step 1: Add markered docs**

Each generated page must include exactly one marker like:

```markdown
<!-- PROGRESS:START kind=agent-queue -->
_Generated content will be written by `go run ./cmd/progress write`._
<!-- PROGRESS:END -->
```

- [ ] **Step 2: Split Phase 1 rows**

Replace the broad Response and Static Fetcher rows with five rows:

```text
Response metadata and selector contract
Response body, text, bytes, and JSON helpers
Static Fetcher method surface over net/http
FetcherSession option merging and cookies
Redirect, timeout, retry, and error taxonomy
```

The first row should use `contract_status: fixture_ready`; the rest should use
`contract_status: draft`.

- [ ] **Step 3: Run generator**

Run:

```sh
go run ./cmd/progress write
```

Expected: each builder-loop page is regenerated.

- [ ] **Step 4: Validate docs**

Run:

```sh
go run ./cmd/progress validate
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
```

Expected: both pass.

- [ ] **Step 5: Commit**

```sh
git add docs/content/building-goscrapling
git commit -m "docs: generate progress work queues"
```

## Task 4: Behavior Atoms And Boundary Docs

**Files:**
- Create: `docs/content/building-goscrapling/architecture_plan/scrapling-behavior-atoms.md`
- Create: `docs/content/building-goscrapling/architecture_plan/boundaries.md`
- Modify: `docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md`
- Modify: `docs/content/building-goscrapling/architecture_plan/upstream-coverage-ledger.md`

- [ ] **Step 1: Add behavior atoms**

Create sections for:

```text
Response metadata and selector behavior
Response body/json helpers
Static Fetcher methods
FetcherSession defaults, cookies, and option merging
Redirect, timeout, retry, and error taxonomy
Browser fetcher contract
Spider request, scheduler, session, follow, and result contracts
```

Each row must include:

```text
Atom, upstream refs, visible contract, Go target, progress row, validation, status
```

- [ ] **Step 2: Add boundary doc**

Add rules for:

```text
compat/scrapling = evidence and explicit compatibility shims
adaptive/fetcher/browser/spider/storage = durable Go architecture
references/Scrapling = ignored study checkout
no donor package names in exported APIs unless they match user-facing Scrapling concepts
```

- [ ] **Step 3: Link docs from feature map and coverage ledger**

Add links to the new behavior atoms and boundaries docs near the top of both
planning docs.

- [ ] **Step 4: Commit**

```sh
git add docs/content/building-goscrapling/architecture_plan
git commit -m "docs: map scrapling behavior atoms"
```

## Task 5: Stronger Drift Tests And Public Docs

**Files:**
- Modify: `progress_docs_test.go`
- Modify: `README.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: Write failing coverage test change**

Update `progress_docs_test.go` to inventory `references/Scrapling/scrapling`
source classes when present. The represented map should include:

```go
map[string]string{
    "parser.py": "`scrapling/parser.py`",
    "cli.py": "`scrapling/cli.py`",
    "core": "`scrapling/core/storage.py`",
    "engines": "`scrapling/engines/static.py`",
    "fetchers": "`scrapling/fetchers/requests.py`",
    "spiders": "`scrapling/spiders/**`",
}
```

The ignored map should include:

```go
map[string]struct{}{
    "__init__.py": {},
    "py.typed": {},
}
```

- [ ] **Step 2: Run root tests and confirm RED or GREEN**

Run:

```sh
go test ./... -count=1
```

Expected: if coverage ledger is already complete, this may pass. If it fails,
fix the ledger, not the test.

- [ ] **Step 3: Update README and AGENTS**

Mention:

```sh
go run ./cmd/progress validate
go run ./cmd/progress write
```

in both files.

- [ ] **Step 4: Run full validation**

Run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress write
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
```

Expected: all pass.

- [ ] **Step 5: Commit**

```sh
git add progress_docs_test.go README.md AGENTS.md docs/content/building-goscrapling
git commit -m "test: guard scrapling port control plane"
```

## Final Verification

Run:

```sh
go test ./... -count=1
go run ./cmd/progress validate
go run ./cmd/progress write
jq empty docs/content/building-goscrapling/architecture_plan/progress.json
git diff --check
git status --short
```

Expected:

- tests pass;
- progress validates;
- progress write is deterministic;
- JSON parses;
- no whitespace errors;
- only intentional committed changes remain.

Then push `main` and update the parent fleet submodule pointer.
