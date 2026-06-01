package selection

import (
	"sort"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model/schema"
)

type Progress = schema.Progress
type Phase = schema.Phase
type Item = schema.Item
type Row = schema.Row
type Status = schema.Status

const (
	StatusComplete   = schema.StatusComplete
	StatusInProgress = schema.StatusInProgress
	StatusPlanned    = schema.StatusPlanned
)

func AgentQueueRows(p *Progress) []Row {
	rows := make([]Row, 0)
	for _, row := range contractRows(p) {
		item := row.Item
		if item.Status == StatusComplete {
			continue
		}
		if item.SliceSize == "umbrella" {
			continue
		}
		if len(item.BlockedBy) > 0 {
			continue
		}
		if !hasBuilderHandoff(item) {
			continue
		}
		rows = append(rows, row)
	}
	sortRows(rows)
	return rows
}

func BlockedRows(p *Progress) []Row {
	rows := make([]Row, 0)
	for _, row := range contractRows(p) {
		if row.Item.Status != StatusComplete && len(row.Item.BlockedBy) > 0 {
			rows = append(rows, row)
		}
	}
	sortRows(rows)
	return rows
}

func UmbrellaRows(p *Progress) []Row {
	rows := make([]Row, 0)
	for _, row := range contractRows(p) {
		if row.Item.Status != StatusComplete && row.Item.SliceSize == "umbrella" {
			rows = append(rows, row)
		}
	}
	sortRows(rows)
	return rows
}

func Rows(p *Progress) []Row {
	if p == nil {
		return nil
	}
	rows := make([]Row, 0)
	for _, phaseKey := range schema.SortedPhaseKeys(p) {
		phase := p.Phases[phaseKey]
		for _, subphaseKey := range schema.SortedSubphaseKeys(phase) {
			subphase := phase.Subphases[subphaseKey]
			for _, item := range subphase.Items {
				rows = append(rows, Row{PhaseKey: phaseKey, SubphaseKey: subphaseKey, Item: item})
			}
		}
	}
	return rows
}

func contractRows(p *Progress) []Row {
	rows := make([]Row, 0)
	for _, row := range Rows(p) {
		if row.Item.Contract == "" {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func hasBuilderHandoff(item Item) bool {
	return item.Contract != "" &&
		item.SliceSize != "" &&
		item.ExecutionOwner != "" &&
		len(item.SourceRefs) > 0 &&
		len(item.ReadyWhen) > 0 &&
		len(item.WriteScope) > 0 &&
		hasTestEvidence(item) &&
		len(item.Acceptance) > 0 &&
		len(item.DoneSignal) > 0
}

func hasTestEvidence(item Item) bool {
	return len(item.TestCommands) > 0 || strings.TrimSpace(item.NoTestRequired) != ""
}

func sortRows(rows []Row) {
	sort.SliceStable(rows, func(i, j int) bool {
		left := rows[i].Item
		right := rows[j].Item
		if priorityRank(left.Priority) != priorityRank(right.Priority) {
			return priorityRank(left.Priority) < priorityRank(right.Priority)
		}
		if statusRank(left.Status) != statusRank(right.Status) {
			return statusRank(left.Status) < statusRank(right.Status)
		}
		if contractStatusRank(left.ContractStatus) != contractStatusRank(right.ContractStatus) {
			return contractStatusRank(left.ContractStatus) < contractStatusRank(right.ContractStatus)
		}
		if rows[i].PhaseKey != rows[j].PhaseKey {
			return rows[i].PhaseKey < rows[j].PhaseKey
		}
		if rows[i].SubphaseKey != rows[j].SubphaseKey {
			return rows[i].SubphaseKey < rows[j].SubphaseKey
		}
		return left.Name < right.Name
	})
}

func priorityRank(priority string) int {
	switch priority {
	case "P0":
		return 0
	case "P1":
		return 1
	case "P2":
		return 2
	case "P3":
		return 3
	case "P4":
		return 4
	default:
		return 5
	}
}

func statusRank(status Status) int {
	switch status {
	case StatusInProgress:
		return 0
	case StatusPlanned:
		return 1
	default:
		return 2
	}
}

func contractStatusRank(status string) int {
	switch status {
	case "fixture_ready":
		return 0
	case "draft":
		return 1
	case "missing":
		return 2
	case "validated":
		return 3
	default:
		return 4
	}
}
