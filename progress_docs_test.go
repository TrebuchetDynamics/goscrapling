package goscrapling

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	progressmodel "github.com/TrebuchetDynamics/goscrapling/internal/progress"
)

func TestProgressLedgerIsBuilderReady(t *testing.T) {
	progress, err := progressmodel.Load(filepath.Join("docs", "content", "building-goscrapling", "architecture_plan", "progress.json"))
	if err != nil {
		t.Fatalf("load progress.json: %v", err)
	}
	if err := progressmodel.Validate(progress); err != nil {
		t.Fatalf("validate progress.json: %v", err)
	}

	stats := progress.Stats()
	if stats.Items.Complete == 0 {
		t.Fatal("progress ledger should include completed foundation rows")
	}
	if stats.Items.Planned == 0 {
		t.Fatal("progress ledger should include planned parity rows")
	}
	if len(progressmodel.AgentQueueRows(progress)) == 0 {
		t.Fatal("progress ledger should expose at least one builder-ready row")
	}
	if len(progressmodel.UmbrellaRows(progress)) == 0 {
		t.Fatal("progress ledger should keep broad future work visible as umbrella cleanup")
	}
}

func TestUpstreamCoverageLedgerClassifiesScraplingSourceClasses(t *testing.T) {
	root := filepath.Join("references", "Scrapling", "scrapling")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("Scrapling checkout not present: %v", err)
	}
	body, err := os.ReadFile(filepath.Join("docs", "content", "building-goscrapling", "architecture_plan", "upstream-coverage-ledger.md"))
	if err != nil {
		t.Fatalf("read coverage ledger: %v", err)
	}
	ledger := string(body)
	represented := map[string]string{
		"parser.py": "`scrapling/parser.py`",
		"cli.py":    "`scrapling/cli.py`",
		"core":      "`scrapling/core/storage.py`",
		"engines":   "`scrapling/engines/static.py`",
		"fetchers":  "`scrapling/fetchers/requests.py`",
		"spiders":   "`scrapling/spiders/**`",
	}
	ignored := map[string]struct{}{
		"__init__.py": {},
		"py.typed":    {},
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read upstream source classes: %v", err)
	}
	var unknown []string
	for _, entry := range entries {
		name := entry.Name()
		if _, ok := ignored[name]; ok {
			continue
		}
		evidence, ok := represented[name]
		if !ok {
			unknown = append(unknown, name)
			continue
		}
		if !strings.Contains(ledger, evidence) {
			t.Errorf("coverage ledger missing evidence %s for upstream source class %s", evidence, name)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		t.Fatalf("coverage ledger has no classification for upstream source classes: %s", strings.Join(unknown, ", "))
	}
}
