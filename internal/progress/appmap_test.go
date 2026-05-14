package progress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAppMapRejectsIncompleteEntries(t *testing.T) {
	err := ValidateAppMap(&AppMap{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	for _, want := range []string{
		"meta.version is required",
		"entries are required",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("validation error missing %q: %v", want, err)
		}
	}

	appMap := fixtureAppMap()
	appMap.Entries = append(appMap.Entries, AppMapEntry{
		ID:                     appMap.Entries[0].ID,
		Title:                  "Incomplete duplicate",
		CoverageStatus:         "unknown",
		TranslationSuitability: "unknown",
		Upstream: []AppMapRef{
			{Ref: "", Kind: "source"},
			{Ref: "references/Scrapling/README.md", Kind: "readme"},
		},
	})

	err = ValidateAppMap(appMap)
	if err == nil {
		t.Fatal("expected entry validation error")
	}
	message = err.Error()
	for _, want := range []string{
		"duplicate entry id",
		"upstream[0] missing ref",
		"invalid upstream kind",
		"invalid coverage_status",
		"invalid translation_suitability",
		"missing feature_anchor",
		"missing go_target",
		"missing progress_rows",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("validation error missing %q: %v", want, err)
		}
	}
}

func TestValidateAppMapCoverageFindsUnmappedUpstreamRefs(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		"references/Scrapling/scrapling/parser.py",
		"references/Scrapling/tests/parser/test_general.py",
		"references/Scrapling/docs/parsing/main_classes.md",
	} {
		writeTempFile(t, root, path)
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].Upstream = appMap.Entries[0].Upstream[:2]
	err := ValidateAppMapCoverage(root, appMap)
	if err == nil {
		t.Fatal("expected unmapped upstream ref error")
	}
	if !strings.Contains(err.Error(), "unmapped upstream ref references/Scrapling/docs/parsing/main_classes.md") {
		t.Fatalf("coverage error missing unmapped ref: %v", err)
	}

	if err := ValidateAppMapCoverage(root, fixtureAppMap()); err != nil {
		t.Fatalf("ValidateAppMapCoverage: %v", err)
	}
}

func TestValidateAppMapCoverageRejectsStaleMappedRefs(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		"references/Scrapling/scrapling/parser.py",
		"references/Scrapling/tests/parser/test_general.py",
		"references/Scrapling/docs/parsing/main_classes.md",
	} {
		writeTempFile(t, root, path)
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].Upstream = append(appMap.Entries[0].Upstream, AppMapRef{
		Ref:  "references/Scrapling/scrapling/missing.py",
		Kind: "source",
	})

	err := ValidateAppMapCoverage(root, appMap)
	if err == nil {
		t.Fatal("expected stale upstream ref error")
	}
	if !strings.Contains(err.Error(), "stale upstream ref references/Scrapling/scrapling/missing.py") {
		t.Fatalf("coverage error missing stale ref: %v", err)
	}
}

func TestValidateAppMapCoverageRejectsStaleMappedDocsAssetMarkdown(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		"references/Scrapling/scrapling/parser.py",
		"references/Scrapling/tests/parser/test_general.py",
		"references/Scrapling/docs/parsing/main_classes.md",
	} {
		writeTempFile(t, root, path)
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].Upstream = append(appMap.Entries[0].Upstream, AppMapRef{
		Ref:  "references/Scrapling/docs/assets/missing.md",
		Kind: "doc",
	})

	err := ValidateAppMapCoverage(root, appMap)
	if err == nil {
		t.Fatal("expected stale docs asset markdown ref error")
	}
	if !strings.Contains(err.Error(), "stale upstream ref references/Scrapling/docs/assets/missing.md") {
		t.Fatalf("coverage error missing stale docs asset markdown ref: %v", err)
	}
}

func TestValidateAppMapCoverageIncludesPyTypedAndDocsAssets(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		"references/Scrapling/scrapling/parser.py",
		"references/Scrapling/scrapling/py.typed",
		"references/Scrapling/tests/parser/test_general.py",
		"references/Scrapling/docs/parsing/main_classes.md",
		"references/Scrapling/docs/assets/logo.svg",
	} {
		writeTempFile(t, root, path)
	}

	err := ValidateAppMapCoverage(root, fixtureAppMap())
	if err == nil {
		t.Fatal("expected unmapped py.typed and docs asset errors")
	}
	message := err.Error()
	for _, want := range []string{
		"unmapped upstream ref references/Scrapling/docs/assets/logo.svg",
		"unmapped upstream ref references/Scrapling/scrapling/py.typed",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("coverage error missing %q: %v", want, err)
		}
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].Upstream = append(appMap.Entries[0].Upstream,
		AppMapRef{Ref: "references/Scrapling/scrapling/py.typed", Kind: "source"},
		AppMapRef{Ref: "references/Scrapling/docs/assets/logo.svg", Kind: "asset"},
	)
	if err := ValidateAppMapCoverage(root, appMap); err != nil {
		t.Fatalf("ValidateAppMapCoverage: %v", err)
	}
}

func TestValidateAppMapCoverageSkipsAbsentUpstreamCheckout(t *testing.T) {
	if err := ValidateAppMapCoverage(t.TempDir(), fixtureAppMap()); err != nil {
		t.Fatalf("ValidateAppMapCoverage: %v", err)
	}
}

func TestValidateAppMapReferencesRejectsUnknownProgressRows(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "docs/research/python-to-go-probes/py2many/go/parser.go.txt")

	appMap := fixtureAppMap()
	appMap.Entries[0].ProgressRows = []string{"Missing row"}
	err := ValidateAppMapReferences(root, appMap, fixtureAppMapProgress())
	if err == nil {
		t.Fatal("expected unknown progress row error")
	}
	if !strings.Contains(err.Error(), `unknown progress row "Missing row"`) {
		t.Fatalf("reference error missing unknown progress row: %v", err)
	}
}

func TestValidateAppMapReferencesRejectsMissingStaticReferencePaths(t *testing.T) {
	err := ValidateAppMapReferences(t.TempDir(), fixtureAppMap(), fixtureAppMapProgress())
	if err == nil {
		t.Fatal("expected missing static reference path error")
	}
	if !strings.Contains(err.Error(), "missing static reference path docs/research/python-to-go-probes/py2many/go/parser.go.txt") {
		t.Fatalf("reference error missing static reference path: %v", err)
	}
}

func TestValidateAppMapReferencesRejectsBlankStaticReferencePaths(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "docs/research/python-to-go-probes/py2many/go/parser.go.txt")

	appMap := fixtureAppMap()
	appMap.Entries[0].StaticReferencePaths = append(appMap.Entries[0].StaticReferencePaths, " \t ")

	err := ValidateAppMapReferences(root, appMap, fixtureAppMapProgress())
	if err == nil {
		t.Fatal("expected blank static reference path error")
	}
	if !strings.Contains(err.Error(), `blank static reference path`) {
		t.Fatalf("reference error missing blank static reference path: %v", err)
	}
}

func TestValidateAppMapReferencesRejectsEscapingStaticReferencePaths(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "docs/research/python-to-go-probes/py2many/go/parser.go.txt")
	outsidePath := filepath.Join(root, "..", "outside-reference.txt")
	if err := os.WriteFile(outsidePath, []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("WriteFile outside reference: %v", err)
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].StaticReferencePaths = append(appMap.Entries[0].StaticReferencePaths, "../outside-reference.txt")

	err := ValidateAppMapReferences(root, appMap, fixtureAppMapProgress())
	if err == nil {
		t.Fatal("expected escaping static reference path error")
	}
	if !strings.Contains(err.Error(), `static reference path escapes repo root ../outside-reference.txt`) {
		t.Fatalf("reference error missing escaping static reference path: %v", err)
	}
}

func TestValidateAppMapReferencesRejectsAbsoluteStaticReferencePaths(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "docs/research/python-to-go-probes/py2many/go/parser.go.txt")
	absolutePath := filepath.Join(root, "absolute-reference.txt")
	if err := os.WriteFile(absolutePath, []byte("absolute\n"), 0o644); err != nil {
		t.Fatalf("WriteFile absolute reference: %v", err)
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].StaticReferencePaths = append(appMap.Entries[0].StaticReferencePaths, absolutePath)

	err := ValidateAppMapReferences(root, appMap, fixtureAppMapProgress())
	if err == nil {
		t.Fatal("expected absolute static reference path error")
	}
	if !strings.Contains(err.Error(), `absolute static reference path `+absolutePath) {
		t.Fatalf("reference error missing absolute static reference path: %v", err)
	}
}

func TestValidateAppMapReferencesRejectsDirectoryStaticReferencePaths(t *testing.T) {
	root := t.TempDir()
	writeTempFile(t, root, "docs/research/python-to-go-probes/py2many/go/parser.go.txt")
	dirPath := filepath.Join(root, "docs", "reference-dir")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("MkdirAll reference dir: %v", err)
	}

	appMap := fixtureAppMap()
	appMap.Entries[0].StaticReferencePaths = append(appMap.Entries[0].StaticReferencePaths, "docs/reference-dir")

	err := ValidateAppMapReferences(root, appMap, fixtureAppMapProgress())
	if err == nil {
		t.Fatal("expected directory static reference path error")
	}
	if !strings.Contains(err.Error(), `static reference path is not a regular file docs/reference-dir`) {
		t.Fatalf("reference error missing directory static reference path: %v", err)
	}
}

func TestRenderAppMapMarkdownIncludesEntries(t *testing.T) {
	body := RenderAppMapMarkdown(fixtureAppMap())

	for _, want := range []string{
		"# Upstream Scrapling App Map",
		"- Upstream name: `Scrapling`",
		"## Coverage Summary",
		"| Status | Count |",
		"| Parser core | `covered` | `parser` | `parser.go` | `Port parser basics` | `manual_rewrite` | 3 |",
		"## Parser core",
		"- `references/Scrapling/scrapling/parser.py` (`source`) symbols: `Adaptor`, `TextHandler`",
		"- Static reference paths: `docs/research/python-to-go-probes/py2many/go/parser.go.txt`",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered markdown missing %q:\n%s", want, body)
		}
	}
}

func fixtureAppMap() *AppMap {
	return &AppMap{
		Meta: AppMapMeta{
			Version: "1.0",
			Upstream: UpstreamMeta{
				Name:            "Scrapling",
				Repo:            "https://github.com/D4Vinci/Scrapling",
				ObservedCommit:  "6380ef0f266a5fff898c18953d6b03ca320b2fd4",
				ObservedRelease: "v0.4.8",
				LocalCheckout:   "references/Scrapling",
			},
			GeneratedMarkdown: "docs/content/building-goscrapling/architecture_plan/upstream-app-map.md",
			Py2ManyProbeDir:   "docs/research/python-to-go-probes/py2many",
		},
		Entries: []AppMapEntry{
			{
				ID:            "parser-core",
				Title:         "Parser core",
				FeatureAnchor: "parser",
				Upstream: []AppMapRef{
					{
						Ref:     "references/Scrapling/scrapling/parser.py",
						Kind:    "source",
						Symbols: []string{"Adaptor", "TextHandler"},
					},
					{Ref: "references/Scrapling/tests/parser/test_general.py", Kind: "test"},
					{Ref: "references/Scrapling/docs/parsing/main_classes.md", Kind: "doc"},
				},
				BehaviorAtoms:          []string{"Text and selector helpers mirror Scrapling parser behavior."},
				GoTarget:               "parser.go",
				ProgressRows:           []string{"Port parser basics"},
				CoverageStatus:         "covered",
				TranslationSuitability: "manual_rewrite",
				StaticReferencePaths:   []string{"docs/research/python-to-go-probes/py2many/go/parser.go.txt"},
				Notes:                  []string{"Reference-only static translation output must not be copied."},
			},
		},
	}
}

func fixtureAppMapProgress() *Progress {
	return &Progress{
		Meta: Meta{Version: "1.0"},
		Phases: map[string]Phase{
			"parser": {
				Name: "Parser",
				Subphases: map[string]Subphase{
					"core": {
						Name: "Core",
						Items: []Item{
							{Name: "Port parser basics", Status: StatusPlanned},
						},
					},
				},
			},
		},
	}
}

func writeTempFile(t *testing.T, root, path string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("fixture\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
