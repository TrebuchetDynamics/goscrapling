package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model"
)

func Validate(p *model.Progress) error {
	var errs []error
	if p == nil {
		return errors.New("progress: nil progress")
	}
	if strings.TrimSpace(p.Meta.Version) == "" {
		errs = append(errs, errors.New("progress: meta.version is required"))
	}
	errs = append(errs, validateBuilderLoopMeta(p.Meta.BuilderLoop)...)
	if len(p.Phases) == 0 {
		errs = append(errs, errors.New("progress: phases are required"))
	}
	for _, phaseKey := range model.SortedPhaseKeys(p) {
		phase := p.Phases[phaseKey]
		if strings.TrimSpace(phase.Name) == "" {
			errs = append(errs, fmt.Errorf("progress: phase %s missing name", phaseKey))
		}
		if len(phase.Subphases) == 0 {
			errs = append(errs, fmt.Errorf("progress: phase %s missing subphases", phaseKey))
			continue
		}
		for _, subphaseKey := range model.SortedSubphaseKeys(phase) {
			subphase := phase.Subphases[subphaseKey]
			if strings.TrimSpace(subphase.Name) == "" {
				errs = append(errs, fmt.Errorf("progress: phase %s subphase %s missing name", phaseKey, subphaseKey))
			}
			if len(subphase.Items) == 0 {
				errs = append(errs, fmt.Errorf("progress: phase %s subphase %s missing items", phaseKey, subphaseKey))
				continue
			}
			for i, item := range subphase.Items {
				errs = append(errs, validateItem(phaseKey, subphaseKey, i, item)...)
			}
		}
	}
	return errors.Join(errs...)
}

func validateBuilderLoopMeta(meta model.BuilderLoopMeta) []error {
	if !builderLoopMetaDeclared(meta) {
		return nil
	}
	var errs []error
	required := []struct {
		name  string
		value string
	}{
		{name: "entrypoint", value: meta.Entrypoint},
		{name: "plan", value: meta.Plan},
		{name: "progress_schema", value: meta.ProgressSchema},
		{name: "candidate_source", value: meta.CandidateSource},
		{name: "unit_test", value: meta.UnitTest},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			errs = append(errs, fmt.Errorf("progress: meta.builder_loop missing %s", field.name))
		}
	}
	if len(meta.CandidatePolicy) == 0 {
		errs = append(errs, errors.New("progress: meta.builder_loop missing candidate_policy"))
	}
	return errs
}

func builderLoopMetaDeclared(meta model.BuilderLoopMeta) bool {
	return meta.Entrypoint != "" ||
		meta.Plan != "" ||
		meta.CoverageLedger != "" ||
		meta.AgentQueue != "" ||
		meta.ProgressSchema != "" ||
		meta.CandidateSource != "" ||
		meta.UnitTest != "" ||
		len(meta.CandidatePolicy) > 0
}

func validateItem(phaseKey, subphaseKey string, index int, item model.Item) []error {
	var errs []error
	prefix := func() string {
		return fmt.Sprintf("progress: phase %s subphase %s item[%d] (%q): ", phaseKey, subphaseKey, index, item.Name)
	}
	add := func(message string) {
		errs = append(errs, errors.New(prefix()+message))
	}
	if strings.TrimSpace(item.Name) == "" {
		add("missing name")
	}
	if !validStatus(item.Status) {
		add(fmt.Sprintf("invalid status %q", item.Status))
	}
	if item.Priority != "" && !validPriority(item.Priority) {
		add(fmt.Sprintf("invalid priority %q", item.Priority))
	}
	if item.ContractStatus != "" && !validContractStatus(item.ContractStatus) {
		add(fmt.Sprintf("invalid contract_status %q", item.ContractStatus))
	}
	if item.SliceSize != "" && !validSliceSize(item.SliceSize) {
		add(fmt.Sprintf("invalid slice_size %q", item.SliceSize))
	}
	if item.ExecutionOwner != "" && !validExecutionOwner(item.ExecutionOwner) {
		add(fmt.Sprintf("invalid execution_owner %q", item.ExecutionOwner))
	}
	if item.Status == model.StatusInProgress && item.SliceSize == "umbrella" {
		add("in_progress row cannot use slice_size umbrella")
	}
	if item.Status == model.StatusComplete && item.Contract != "" && item.ContractStatus != "validated" {
		add("complete contract row must use contract_status validated")
	}
	if item.Contract != "" {
		errs = append(errs, validateContractMetadata(prefix, item)...)
	}
	return errs
}

func validateContractMetadata(prefix func() string, item model.Item) []error {
	var errs []error
	add := func(message string) {
		errs = append(errs, errors.New(prefix()+message))
	}
	if item.ContractStatus == "" {
		add("missing contract_status")
	}
	if item.SliceSize == "" {
		add("missing slice_size")
	}
	if item.ExecutionOwner == "" {
		add("missing execution_owner")
	}
	if len(item.SourceRefs) == 0 {
		add("missing source_refs")
	}
	if len(item.ReadyWhen) == 0 {
		add("missing ready_when")
	}
	if len(item.WriteScope) == 0 {
		add("missing write_scope")
	}
	if len(item.TestCommands) == 0 && strings.TrimSpace(item.NoTestRequired) == "" {
		add("missing test_commands or no_test_required")
	}
	if len(item.Acceptance) == 0 {
		add("missing acceptance")
	}
	if len(item.DoneSignal) == 0 {
		add("missing done_signal")
	}
	if len(item.BlockedBy) > 0 && len(item.ReadyWhen) == 0 {
		add("blocked row missing ready_when")
	}
	return errs
}

func validStatus(status model.Status) bool {
	return status == model.StatusComplete || status == model.StatusInProgress || status == model.StatusPlanned
}

func validPriority(priority string) bool {
	switch priority {
	case "P0", "P1", "P2", "P3", "P4":
		return true
	default:
		return false
	}
}

func validContractStatus(status string) bool {
	switch status {
	case "missing", "draft", "fixture_ready", "validated":
		return true
	default:
		return false
	}
}

func validSliceSize(size string) bool {
	switch size {
	case "small", "medium", "large", "umbrella":
		return true
	default:
		return false
	}
}

func validExecutionOwner(owner string) bool {
	switch owner {
	case "parser", "storage", "fetcher", "browser", "spider", "cli", "integration", "docs":
		return true
	default:
		return false
	}
}
