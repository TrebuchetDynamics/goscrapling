package goscrapling

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type progressDocument struct {
	Meta struct {
		Version     string `json:"version"`
		BuilderLoop struct {
			Entrypoint      string   `json:"entrypoint"`
			Plan            string   `json:"plan"`
			CandidateSource string   `json:"candidate_source"`
			UnitTest        string   `json:"unit_test"`
			CandidatePolicy []string `json:"candidate_policy"`
		} `json:"builder_loop"`
	} `json:"meta"`
	Phases map[string]progressPhase `json:"phases"`
}

type progressPhase struct {
	Name      string                      `json:"name"`
	Subphases map[string]progressSubphase `json:"subphases"`
}

type progressSubphase struct {
	Name  string         `json:"name"`
	Items []progressItem `json:"items"`
}

type progressItem struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Priority       string   `json:"priority"`
	Contract       string   `json:"contract"`
	ContractStatus string   `json:"contract_status"`
	SliceSize      string   `json:"slice_size"`
	ExecutionOwner string   `json:"execution_owner"`
	SourceRefs     []string `json:"source_refs"`
	ReadyWhen      []string `json:"ready_when"`
	NotReadyWhen   []string `json:"not_ready_when"`
	WriteScope     []string `json:"write_scope"`
	TestCommands   []string `json:"test_commands"`
	NoTestRequired string   `json:"no_test_required"`
	Acceptance     []string `json:"acceptance"`
	DoneSignal     []string `json:"done_signal"`
}

func TestProgressLedgerIsBuilderReady(t *testing.T) {
	progress := readProgressDocument(t)

	if progress.Meta.Version == "" {
		t.Fatal("progress meta.version is required")
	}
	if progress.Meta.BuilderLoop.Entrypoint == "" {
		t.Fatal("progress meta.builder_loop.entrypoint is required")
	}
	if progress.Meta.BuilderLoop.Plan == "" {
		t.Fatal("progress meta.builder_loop.plan is required")
	}
	if progress.Meta.BuilderLoop.CandidateSource == "" {
		t.Fatal("progress meta.builder_loop.candidate_source is required")
	}
	if progress.Meta.BuilderLoop.UnitTest == "" {
		t.Fatal("progress meta.builder_loop.unit_test is required")
	}
	if len(progress.Meta.BuilderLoop.CandidatePolicy) == 0 {
		t.Fatal("progress meta.builder_loop.candidate_policy is required")
	}
	if len(progress.Phases) == 0 {
		t.Fatal("progress phases are required")
	}

	statuses := map[string]bool{
		"planned":     true,
		"in_progress": true,
		"complete":    true,
	}
	contractStatuses := map[string]bool{
		"missing":       true,
		"draft":         true,
		"fixture_ready": true,
		"validated":     true,
	}
	var completeCount, plannedCount int
	for phaseKey, phase := range progress.Phases {
		if phase.Name == "" {
			t.Fatalf("phase %s missing name", phaseKey)
		}
		if len(phase.Subphases) == 0 {
			t.Fatalf("phase %s missing subphases", phaseKey)
		}
		for subphaseKey, subphase := range phase.Subphases {
			if subphase.Name == "" {
				t.Fatalf("phase %s subphase %s missing name", phaseKey, subphaseKey)
			}
			if len(subphase.Items) == 0 {
				t.Fatalf("phase %s subphase %s missing items", phaseKey, subphaseKey)
			}
			for i, item := range subphase.Items {
				path := phaseKey + "/" + subphaseKey + "/" + item.Name
				if item.Name == "" {
					t.Fatalf("phase %s subphase %s item[%d] missing name", phaseKey, subphaseKey, i)
				}
				if !statuses[item.Status] {
					t.Fatalf("%s has invalid status %q", path, item.Status)
				}
				if item.Status == "complete" {
					completeCount++
				}
				if item.Status == "planned" {
					plannedCount++
				}
				if item.Contract == "" {
					t.Fatalf("%s missing contract", path)
				}
				if !contractStatuses[item.ContractStatus] {
					t.Fatalf("%s has invalid contract_status %q", path, item.ContractStatus)
				}
				if item.Status == "complete" && item.ContractStatus != "validated" {
					t.Fatalf("%s is complete but contract_status is %q", path, item.ContractStatus)
				}
				if item.SliceSize == "" {
					t.Fatalf("%s missing slice_size", path)
				}
				if item.Status == "in_progress" && item.SliceSize == "umbrella" {
					t.Fatalf("%s is in_progress but still an umbrella row", path)
				}
				if item.ExecutionOwner == "" {
					t.Fatalf("%s missing execution_owner", path)
				}
				if len(item.SourceRefs) == 0 {
					t.Fatalf("%s missing source_refs", path)
				}
				if len(item.ReadyWhen) == 0 {
					t.Fatalf("%s missing ready_when", path)
				}
				if len(item.WriteScope) == 0 {
					t.Fatalf("%s missing write_scope", path)
				}
				if len(item.TestCommands) == 0 && strings.TrimSpace(item.NoTestRequired) == "" {
					t.Fatalf("%s missing test_commands or no_test_required", path)
				}
				if len(item.Acceptance) == 0 {
					t.Fatalf("%s missing acceptance", path)
				}
				if len(item.DoneSignal) == 0 {
					t.Fatalf("%s missing done_signal", path)
				}
			}
		}
	}
	if completeCount == 0 {
		t.Fatal("progress ledger should include completed foundation rows")
	}
	if plannedCount == 0 {
		t.Fatal("progress ledger should include planned parity rows")
	}
}

func TestUpstreamCoverageLedgerMentionsCoreScraplingSources(t *testing.T) {
	if _, err := os.Stat(filepath.Join("references", "Scrapling", "scrapling")); err != nil {
		t.Skipf("Scrapling checkout not present: %v", err)
	}
	body, err := os.ReadFile(filepath.Join("docs", "content", "building-goscrapling", "architecture_plan", "upstream-coverage-ledger.md"))
	if err != nil {
		t.Fatalf("read coverage ledger: %v", err)
	}
	ledger := string(body)
	required := []string{
		"scrapling/parser.py",
		"scrapling/core/storage.py",
		"scrapling/engines/static.py",
		"scrapling/fetchers/requests.py",
		"scrapling/fetchers/chrome.py",
		"scrapling/fetchers/stealth_chrome.py",
		"scrapling/spiders/**",
		"scrapling/cli.py",
	}
	for _, source := range required {
		if !strings.Contains(ledger, source) {
			t.Fatalf("coverage ledger missing %s", source)
		}
	}
}

func readProgressDocument(t *testing.T) progressDocument {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("docs", "content", "building-goscrapling", "architecture_plan", "progress.json"))
	if err != nil {
		t.Fatalf("read progress.json: %v", err)
	}
	var progress progressDocument
	if err := json.Unmarshal(body, &progress); err != nil {
		t.Fatalf("parse progress.json: %v", err)
	}
	return progress
}
