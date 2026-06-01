package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/schema"
)

func TestValidateCoverageRejectsUnsupportedMappedUpstreamRefs(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "references/Scrapling/scrapling/parser.py")

	appMap := &schema.AppMap{Entries: []schema.AppMapEntry{{
		ID: "unsupported-ref",
		Upstream: []schema.AppMapRef{
			{Ref: "references/Scrapling/scrapling/parser.py", Kind: "source"},
			{Ref: "references/Scrapling/README.md", Kind: "doc"},
		},
	}}}

	err := ValidateCoverage(root, appMap)
	if err == nil {
		t.Fatal("expected unsupported upstream ref error")
	}
	if !strings.Contains(err.Error(), "unsupported upstream ref references/Scrapling/README.md") {
		t.Fatalf("coverage error missing unsupported ref: %v", err)
	}
}

func TestValidateCoverageRejectsEscapingMappedUpstreamRefs(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "references/Scrapling/scrapling/parser.py")

	appMap := &schema.AppMap{Entries: []schema.AppMapEntry{{
		ID: "escaping-ref",
		Upstream: []schema.AppMapRef{
			{Ref: "../Scrapling/scrapling/parser.py", Kind: "source"},
		},
	}}}

	err := ValidateCoverage(root, appMap)
	if err == nil {
		t.Fatal("expected escaping upstream ref error")
	}
	if !strings.Contains(err.Error(), "upstream ref escapes repo root ../Scrapling/scrapling/parser.py") {
		t.Fatalf("coverage error missing escaping ref: %v", err)
	}
}

func writeFixtureFile(t *testing.T, root, path string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("fixture\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
