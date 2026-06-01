package model

import (
	"reflect"
	"testing"
)

func TestRowsReturnsSortedPhaseAndSubphaseRows(t *testing.T) {
	progress := &Progress{Phases: map[string]Phase{
		"phase-b": {Subphases: map[string]Subphase{
			"sub-b": {Items: []Item{{Name: "third", Status: StatusPlanned}}},
		}},
		"phase-a": {Subphases: map[string]Subphase{
			"sub-b": {Items: []Item{{Name: "second", Status: StatusPlanned}}},
			"sub-a": {Items: []Item{{Name: "first", Status: StatusPlanned}}},
		}},
	}}

	rows := Rows(progress)
	got := make([]string, 0, len(rows))
	for _, row := range rows {
		got = append(got, row.PhaseKey+"/"+row.SubphaseKey+"/"+row.Item.Name)
	}
	want := []string{"phase-a/sub-a/first", "phase-a/sub-b/second", "phase-b/sub-b/third"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Rows order = %#v, want %#v", got, want)
	}
}

func TestAgentQueueRowsFiltersAndSortsBuilderReadyRows(t *testing.T) {
	progress := &Progress{Phases: map[string]Phase{
		"phase": {Subphases: map[string]Subphase{
			"subphase": {Items: []Item{
				builderReadyRow("P2 row", "P2", StatusPlanned),
				builderReadyRow("P0 row", "P0", StatusPlanned),
				builderReadyRow("In progress row", "P1", StatusInProgress),
				builderReadyRow("Complete row", "P0", StatusComplete),
				func() Item {
					item := builderReadyRow("Blocked row", "P0", StatusPlanned)
					item.BlockedBy = []string{"dependency"}
					return item
				}(),
				func() Item {
					item := builderReadyRow("Umbrella row", "P0", StatusPlanned)
					item.SliceSize = "umbrella"
					return item
				}(),
				{Name: "No contract", Priority: "P0", Status: StatusPlanned},
			}},
		}},
	}}

	rows := AgentQueueRows(progress)
	got := make([]string, 0, len(rows))
	for _, row := range rows {
		got = append(got, row.Item.Name)
	}
	want := []string{"P0 row", "In progress row", "P2 row"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AgentQueueRows = %#v, want %#v", got, want)
	}
}

func TestAgentQueueRowsRejectsWhitespaceOnlyNoTestReason(t *testing.T) {
	progress := &Progress{Phases: map[string]Phase{
		"phase": {Subphases: map[string]Subphase{
			"subphase": {Items: []Item{
				func() Item {
					item := builderReadyRow("Whitespace no-test row", "P0", StatusPlanned)
					item.TestCommands = nil
					item.NoTestRequired = "   "
					return item
				}(),
			}},
		}},
	}}

	rows := AgentQueueRows(progress)
	if len(rows) != 0 {
		t.Fatalf("AgentQueueRows included whitespace-only no_test_required row: %#v", rows)
	}
}

func TestStatsDerivesAggregateStatuses(t *testing.T) {
	progress := &Progress{Phases: map[string]Phase{
		"done": {Subphases: map[string]Subphase{
			"done-sub": {Items: []Item{{Status: StatusComplete}}},
		}},
		"active": {Subphases: map[string]Subphase{
			"active-sub": {Items: []Item{{Status: StatusComplete}, {Status: StatusPlanned}}},
		}},
		"planned": {Subphases: map[string]Subphase{
			"planned-sub": {Items: []Item{{Status: StatusPlanned}}},
		}},
	}}

	stats := progress.Stats()
	if stats.Phases != (Counts{Total: 3, Complete: 1, InProgress: 1, Planned: 1}) {
		t.Fatalf("phase stats = %#v", stats.Phases)
	}
	if stats.Subphases != (Counts{Total: 3, Complete: 1, InProgress: 1, Planned: 1}) {
		t.Fatalf("subphase stats = %#v", stats.Subphases)
	}
	if stats.Items != (Counts{Total: 4, Complete: 2, Planned: 2}) {
		t.Fatalf("item stats = %#v", stats.Items)
	}
}

func builderReadyRow(name, priority string, status Status) Item {
	contractStatus := "draft"
	if status == StatusComplete {
		contractStatus = "validated"
	}
	return Item{
		Name:           name,
		Priority:       priority,
		Status:         status,
		Contract:       "contract",
		ContractStatus: contractStatus,
		SliceSize:      "small",
		ExecutionOwner: "agent",
		SourceRefs:     []string{"source"},
		ReadyWhen:      []string{"ready"},
		WriteScope:     []string{"scope"},
		TestCommands:   []string{"go test ./..."},
		Acceptance:     []string{"accepted"},
		DoneSignal:     []string{"done"},
	}
}
