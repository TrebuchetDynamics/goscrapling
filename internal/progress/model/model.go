package model

import (
	"encoding/json"
	"os"
	"sort"
)

type Status string

const (
	StatusComplete   Status = "complete"
	StatusInProgress Status = "in_progress"
	StatusPlanned    Status = "planned"
)

type Meta struct {
	Version     string          `json:"version"`
	PortTarget  string          `json:"port_target,omitempty"`
	Upstream    UpstreamMeta    `json:"upstream,omitempty"`
	BuilderLoop BuilderLoopMeta `json:"builder_loop,omitempty"`
}

type UpstreamMeta struct {
	Name            string `json:"name,omitempty"`
	Repo            string `json:"repo,omitempty"`
	ObservedCommit  string `json:"observed_commit,omitempty"`
	ObservedRelease string `json:"observed_release,omitempty"`
	LocalCheckout   string `json:"local_checkout,omitempty"`
}

type BuilderLoopMeta struct {
	Entrypoint      string   `json:"entrypoint,omitempty"`
	Plan            string   `json:"plan,omitempty"`
	CoverageLedger  string   `json:"coverage_ledger,omitempty"`
	AgentQueue      string   `json:"agent_queue,omitempty"`
	ProgressSchema  string   `json:"progress_schema,omitempty"`
	CandidateSource string   `json:"candidate_source,omitempty"`
	UnitTest        string   `json:"unit_test,omitempty"`
	CandidatePolicy []string `json:"candidate_policy,omitempty"`
}

type Progress struct {
	Meta   Meta             `json:"meta"`
	Phases map[string]Phase `json:"phases"`
}

type Phase struct {
	Name      string              `json:"name"`
	Subphases map[string]Subphase `json:"subphases"`
}

type Subphase struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

type Item struct {
	Name           string   `json:"name"`
	Priority       string   `json:"priority,omitempty"`
	Status         Status   `json:"status"`
	Contract       string   `json:"contract,omitempty"`
	ContractStatus string   `json:"contract_status,omitempty"`
	SliceSize      string   `json:"slice_size,omitempty"`
	ExecutionOwner string   `json:"execution_owner,omitempty"`
	SourceRefs     []string `json:"source_refs,omitempty"`
	ReadyWhen      []string `json:"ready_when,omitempty"`
	NotReadyWhen   []string `json:"not_ready_when,omitempty"`
	BlockedBy      []string `json:"blocked_by,omitempty"`
	Unblocks       []string `json:"unblocks,omitempty"`
	WriteScope     []string `json:"write_scope,omitempty"`
	TestCommands   []string `json:"test_commands,omitempty"`
	NoTestRequired string   `json:"no_test_required,omitempty"`
	Acceptance     []string `json:"acceptance,omitempty"`
	DoneSignal     []string `json:"done_signal,omitempty"`
}

type Row struct {
	PhaseKey    string
	SubphaseKey string
	Item        Item
}

type Counts struct {
	Total      int
	Complete   int
	InProgress int
	Planned    int
}

type Stats struct {
	Phases    Counts
	Subphases Counts
	Items     Counts
}

func Load(path string) (*Progress, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var progress Progress
	if err := json.Unmarshal(body, &progress); err != nil {
		return nil, err
	}
	return &progress, nil
}

func (p *Progress) Stats() Stats {
	var stats Stats
	stats.Phases.Total = len(p.Phases)
	for _, phaseKey := range SortedPhaseKeys(p) {
		phase := p.Phases[phaseKey]
		phaseStatus := phase.derivedStatus()
		increment(&stats.Phases, phaseStatus)
		stats.Subphases.Total += len(phase.Subphases)
		for _, subphaseKey := range SortedSubphaseKeys(phase) {
			subphase := phase.Subphases[subphaseKey]
			increment(&stats.Subphases, subphase.derivedStatus())
			stats.Items.Total += len(subphase.Items)
			for _, item := range subphase.Items {
				increment(&stats.Items, item.Status)
			}
		}
	}
	return stats
}

func (phase Phase) derivedStatus() Status {
	if len(phase.Subphases) == 0 {
		return StatusPlanned
	}
	allComplete := true
	anyStarted := false
	for _, subphase := range phase.Subphases {
		status := subphase.derivedStatus()
		if status != StatusComplete {
			allComplete = false
		}
		if status == StatusComplete || status == StatusInProgress {
			anyStarted = true
		}
	}
	switch {
	case allComplete:
		return StatusComplete
	case anyStarted:
		return StatusInProgress
	default:
		return StatusPlanned
	}
}

func (subphase Subphase) derivedStatus() Status {
	if len(subphase.Items) == 0 {
		return StatusPlanned
	}
	allComplete := true
	anyStarted := false
	for _, item := range subphase.Items {
		if item.Status != StatusComplete {
			allComplete = false
		}
		if item.Status == StatusComplete || item.Status == StatusInProgress {
			anyStarted = true
		}
	}
	switch {
	case allComplete:
		return StatusComplete
	case anyStarted:
		return StatusInProgress
	default:
		return StatusPlanned
	}
}

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

func contractRows(p *Progress) []Row {
	rows := make([]Row, 0)
	for _, phaseKey := range SortedPhaseKeys(p) {
		phase := p.Phases[phaseKey]
		for _, subphaseKey := range SortedSubphaseKeys(phase) {
			subphase := phase.Subphases[subphaseKey]
			for _, item := range subphase.Items {
				if item.Contract == "" {
					continue
				}
				rows = append(rows, Row{PhaseKey: phaseKey, SubphaseKey: subphaseKey, Item: item})
			}
		}
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
		(len(item.TestCommands) > 0 || item.NoTestRequired != "") &&
		len(item.Acceptance) > 0 &&
		len(item.DoneSignal) > 0
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

func SortedPhaseKeys(p *Progress) []string {
	keys := make([]string, 0, len(p.Phases))
	for key := range p.Phases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func SortedSubphaseKeys(phase Phase) []string {
	keys := make([]string, 0, len(phase.Subphases))
	for key := range phase.Subphases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func increment(counts *Counts, status Status) {
	switch status {
	case StatusComplete:
		counts.Complete++
	case StatusInProgress:
		counts.InProgress++
	default:
		counts.Planned++
	}
}
