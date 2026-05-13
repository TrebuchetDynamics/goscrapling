package progress

import (
	"strings"
	"testing"
)

func TestValidateAcceptsCurrentProgress(t *testing.T) {
	progress, err := Load("../../docs/content/building-goscrapling/architecture_plan/progress.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := Validate(progress); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateRejectsIncompleteContractRows(t *testing.T) {
	progress := &Progress{
		Meta: Meta{Version: "1.0"},
		Phases: map[string]Phase{
			"phase": {
				Name: "Phase",
				Subphases: map[string]Subphase{
					"subphase": {
						Name: "Subphase",
						Items: []Item{
							{
								Name:           "Broken row",
								Status:         StatusPlanned,
								Contract:       "Port behavior without enough metadata.",
								ContractStatus: "draft",
								SliceSize:      "medium",
							},
						},
					},
				},
			},
		},
	}

	err := Validate(progress)
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	for _, want := range []string{"missing execution_owner", "missing source_refs", "missing write_scope", "missing test_commands or no_test_required", "missing acceptance", "missing done_signal"} {
		if !strings.Contains(message, want) {
			t.Fatalf("validation error missing %q: %v", want, err)
		}
	}
}

func TestAgentQueueExcludesCompleteBlockedAndUmbrellaRows(t *testing.T) {
	progress := fixtureProgress()

	rows := AgentQueueRows(progress)
	if len(rows) != 1 {
		t.Fatalf("agent queue rows = %d, want 1", len(rows))
	}
	if got := rows[0].Item.Name; got != "Ready row" {
		t.Fatalf("agent queue row = %q, want Ready row", got)
	}
}

func TestRenderUmbrellaCleanupIncludesUmbrellaRows(t *testing.T) {
	progress := fixtureProgress()

	body := RenderUmbrellaCleanup(progress)
	if !strings.Contains(body, "Umbrella row") {
		t.Fatalf("umbrella cleanup missing row: %s", body)
	}
	if strings.Contains(body, "Ready row") {
		t.Fatalf("umbrella cleanup included ready row: %s", body)
	}
}

func TestReplaceMarkerReplacesMatchingSection(t *testing.T) {
	input := strings.Join([]string{
		"# Agent Queue",
		"",
		"<!-- PROGRESS:START kind=agent-queue -->",
		"old content",
		"<!-- PROGRESS:END -->",
		"",
	}, "\n")

	out, err := ReplaceMarker(input, "agent-queue", "new content\n")
	if err != nil {
		t.Fatalf("ReplaceMarker: %v", err)
	}
	if strings.Contains(out, "old content") {
		t.Fatalf("old content still present: %s", out)
	}
	if !strings.Contains(out, "new content") {
		t.Fatalf("new content missing: %s", out)
	}
}

func fixtureProgress() *Progress {
	return &Progress{
		Meta: Meta{
			Version: "1.0",
			BuilderLoop: BuilderLoopMeta{
				Entrypoint:      "docs/development-skills/goscrapling-skill-manager/SKILL.md",
				Plan:            "docs/content/building-goscrapling/architecture_plan/scrapling-feature-map.md",
				AgentQueue:      "docs/content/building-goscrapling/builder-loop/agent-queue.md",
				ProgressSchema:  "docs/content/building-goscrapling/builder-loop/progress-schema.md",
				CandidateSource: "docs/content/building-goscrapling/architecture_plan/progress.json",
				UnitTest:        "go test ./... -count=1",
				CandidatePolicy: []string{"Use builder-ready rows."},
			},
		},
		Phases: map[string]Phase{
			"phase": {
				Name: "Phase",
				Subphases: map[string]Subphase{
					"subphase": {
						Name: "Subphase",
						Items: []Item{
							contractRow("Ready row", StatusPlanned, "small"),
							contractRow("Complete row", StatusComplete, "small"),
							contractRow("Umbrella row", StatusPlanned, "umbrella"),
							func() Item {
								item := contractRow("Blocked row", StatusPlanned, "small")
								item.BlockedBy = []string{"Ready row"}
								return item
							}(),
						},
					},
				},
			},
		},
	}
}

func contractRow(name string, status Status, size string) Item {
	contractStatus := "draft"
	if status == StatusComplete {
		contractStatus = "validated"
	}
	return Item{
		Name:           name,
		Status:         status,
		Priority:       "P1",
		Contract:       "A test contract.",
		ContractStatus: contractStatus,
		SliceSize:      size,
		ExecutionOwner: "docs",
		SourceRefs:     []string{"source.md"},
		ReadyWhen:      []string{"ready"},
		NotReadyWhen:   []string{"not ready"},
		WriteScope:     []string{"docs/"},
		TestCommands:   []string{"go test ./..."},
		Acceptance:     []string{"accepted"},
		DoneSignal:     []string{"done"},
	}
}
