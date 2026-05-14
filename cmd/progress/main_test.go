package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidate(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := run(&stdout, &stderr, []string{"--repo-root", "../..", "validate"})
	if err != nil {
		t.Fatalf("run validate: %v\nstderr: %s", err, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "progress: validated") {
		t.Fatalf("stdout = %q, want validation summary", got)
	}
}

func TestRunWriteRegeneratesMarkedDocs(t *testing.T) {
	root := t.TempDir()
	copyFile(t,
		filepath.Join("..", "..", "docs", "content", "building-goscrapling", "architecture_plan", "progress.json"),
		filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "progress.json"),
	)
	builderLoop := filepath.Join(root, "docs", "content", "building-goscrapling", "builder-loop")
	files := map[string]string{
		"builder-loop-handoff.md": "builder-loop-handoff",
		"agent-queue.md":          "agent-queue",
		"next-slices.md":          "next-slices",
		"blocked-slices.md":       "blocked-slices",
		"umbrella-cleanup.md":     "umbrella-cleanup",
	}
	for name, kind := range files {
		writeMarkerFile(t, filepath.Join(builderLoop, name), kind)
	}

	var stdout, stderr bytes.Buffer
	err := run(&stdout, &stderr, []string{"--repo-root", root, "write"})
	if err != nil {
		t.Fatalf("run write: %v\nstderr: %s", err, stderr.String())
	}

	agentQueue, err := os.ReadFile(filepath.Join(builderLoop, "agent-queue.md"))
	if err != nil {
		t.Fatalf("read generated agent queue: %v", err)
	}
	if strings.Contains(string(agentQueue), "old content") {
		t.Fatalf("agent queue still contains old content: %s", agentQueue)
	}
	if !strings.Contains(string(agentQueue), "- Phase:") {
		t.Fatalf("agent queue missing generated row metadata: %s", agentQueue)
	}
}

func TestRunMapValidate(t *testing.T) {
	root := t.TempDir()
	writeMinimalAppMapFixture(t, root)

	var stdout, stderr bytes.Buffer
	err := run(&stdout, &stderr, []string{"--repo-root", root, "map-validate"})
	if err != nil {
		t.Fatalf("run map-validate: %v\nstderr: %s", err, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "app-map: validated") {
		t.Fatalf("stdout = %q, want app-map validation summary", got)
	}
}

func TestRunMapValidateRejectsUnknownProgressRow(t *testing.T) {
	root := t.TempDir()
	writeMinimalAppMapFixture(t, root)
	appMapPath := filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "upstream-app-map.json")
	body, err := os.ReadFile(appMapPath)
	if err != nil {
		t.Fatalf("read app map fixture: %v", err)
	}
	body = []byte(strings.Replace(string(body), `"Port parser basics"`, `"Missing row"`, 1))
	if err := os.WriteFile(appMapPath, body, 0o644); err != nil {
		t.Fatalf("write app map fixture: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = run(&stdout, &stderr, []string{"--repo-root", root, "map-validate"})
	if err == nil {
		t.Fatal("expected map-validate error")
	}
	if !strings.Contains(err.Error(), `unknown progress row "Missing row"`) {
		t.Fatalf("error missing unknown progress row: %v\nstderr: %s", err, stderr.String())
	}
}

func TestRunMapWriteRegeneratesMarkdown(t *testing.T) {
	root := t.TempDir()
	writeMinimalAppMapFixture(t, root)

	var stdout, stderr bytes.Buffer
	err := run(&stdout, &stderr, []string{"--repo-root", root, "map-write"})
	if err != nil {
		t.Fatalf("run map-write: %v\nstderr: %s", err, stderr.String())
	}

	body, err := os.ReadFile(filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "upstream-app-map.md"))
	if err != nil {
		t.Fatalf("read generated app map markdown: %v", err)
	}
	if !strings.Contains(string(body), "# Upstream Scrapling App Map") {
		t.Fatalf("generated app map markdown missing title: %s", body)
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	body, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
	}
	if err := os.WriteFile(dst, body, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func writeMarkerFile(t *testing.T, path, kind string) {
	t.Helper()
	body := "# " + kind + "\n\n" +
		"<!-- PROGRESS:START kind=" + kind + " -->\n" +
		"old content\n" +
		"<!-- PROGRESS:END -->\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeMinimalAppMapFixture(t *testing.T, root string) {
	t.Helper()
	parserPath := filepath.Join(root, "references", "Scrapling", "scrapling", "parser.py")
	if err := os.MkdirAll(filepath.Dir(parserPath), 0o755); err != nil {
		t.Fatalf("mkdir parser fixture: %v", err)
	}
	if err := os.WriteFile(parserPath, []byte("# parser fixture\n"), 0o644); err != nil {
		t.Fatalf("write parser fixture: %v", err)
	}

	appMapPath := filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "upstream-app-map.json")
	if err := os.MkdirAll(filepath.Dir(appMapPath), 0o755); err != nil {
		t.Fatalf("mkdir app map fixture: %v", err)
	}
	body := `{
  "meta": {
    "version": "1.0",
    "upstream": {
      "name": "Scrapling",
      "repo": "https://github.com/D4Vinci/Scrapling",
      "observed_commit": "6380ef0f266a5fff898c18953d6b03ca320b2fd4",
      "observed_release": "v0.4.8",
      "local_checkout": "references/Scrapling"
    },
    "generated_markdown": "docs/content/building-goscrapling/architecture_plan/upstream-app-map.md"
  },
  "entries": [
    {
      "id": "parser-core",
      "title": "Parser core",
      "feature_anchor": "parser",
      "upstream": [
        {
          "ref": "references/Scrapling/scrapling/parser.py",
          "kind": "source",
          "symbols": ["Adaptor"]
        }
      ],
      "go_target": "parser.go",
      "progress_rows": ["Port parser basics"],
      "coverage_status": "covered",
      "translation_suitability": "manual_rewrite"
    }
  ]
}
`
	if err := os.WriteFile(appMapPath, []byte(body), 0o644); err != nil {
		t.Fatalf("write app map fixture: %v", err)
	}

	progressPath := filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "progress.json")
	progressBody := `{
  "meta": {
    "version": "1.0"
  },
  "phases": {
    "parser": {
      "name": "Parser",
      "subphases": {
        "core": {
          "name": "Core",
          "items": [
            {
              "name": "Port parser basics",
              "status": "planned"
            }
          ]
        }
      }
    }
  }
}
`
	if err := os.WriteFile(progressPath, []byte(progressBody), 0o644); err != nil {
		t.Fatalf("write progress fixture: %v", err)
	}
}
