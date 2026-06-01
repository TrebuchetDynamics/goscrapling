package model

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
