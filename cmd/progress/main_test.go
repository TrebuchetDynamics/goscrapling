package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	surfaces := filepath.Join(root, "docs", "content", "building-goscrapling", "builder-loop", "surfaces")
	files := map[string]string{
		filepath.Join("handoff", "builder-loop-handoff.md"):    "builder-loop-handoff",
		filepath.Join("queue", "assignable", "agent-queue.md"): "agent-queue",
		filepath.Join("queue", "assignable", "next-slices.md"): "next-slices",
		filepath.Join("queue", "blocked", "blocked-slices.md"): "blocked-slices",
		filepath.Join("queue", "agent-queue.md"):               "agent-queue",
		filepath.Join("queue", "next-slices.md"):               "next-slices",
		filepath.Join("queue", "blocked-slices.md"):            "blocked-slices",
		filepath.Join("cleanup", "umbrella-cleanup.md"):        "umbrella-cleanup",
	}
	for name, kind := range files {
		writeMarkerFile(t, filepath.Join(surfaces, name), kind)
	}

	var stdout, stderr bytes.Buffer
	err := run(&stdout, &stderr, []string{"--repo-root", root, "write"})
	if err != nil {
		t.Fatalf("run write: %v\nstderr: %s", err, stderr.String())
	}

	for _, path := range []string{
		filepath.Join(surfaces, "queue", "assignable", "agent-queue.md"),
		filepath.Join(surfaces, "queue", "agent-queue.md"),
	} {
		agentQueue, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated agent queue %s: %v", path, err)
		}
		if strings.Contains(string(agentQueue), "old content") {
			t.Fatalf("agent queue still contains old content: %s", agentQueue)
		}
		if !strings.Contains(string(agentQueue), "- Phase:") {
			t.Fatalf("agent queue missing generated row metadata: %s", agentQueue)
		}
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

func TestParityScorecard(t *testing.T) {
	root := t.TempDir()
	copyFile(t,
		filepath.Join("..", "..", "docs", "content", "building-goscrapling", "architecture_plan", "progress.json"),
		filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "progress.json"),
	)
	benchmarkPath := filepath.Join(root, "benchmarks", "parity_bench_test.go")
	if err := os.MkdirAll(filepath.Dir(benchmarkPath), 0o755); err != nil {
		t.Fatalf("mkdir benchmark fixture: %v", err)
	}
	if err := os.WriteFile(benchmarkPath, []byte(`package benchmarks

import "testing"

func BenchmarkParserNestedText(b *testing.B) {}
func BenchmarkStaticFetcherLocalResponse(b *testing.B) {}
func BenchmarkSpiderSchedulerFingerprint(b *testing.B) {}
func BenchmarkCLIExtractFixture(b *testing.B) {}
`), 0o644); err != nil {
		t.Fatalf("write benchmark fixture: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run(&stdout, &stderr, []string{"--repo-root", root, "scorecard"})
	if err != nil {
		t.Fatalf("run scorecard: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "progress: wrote docs/research/parity-scorecard.md") {
		t.Fatalf("stdout = %q, want scorecard write summary", stdout.String())
	}

	body, err := os.ReadFile(filepath.Join(root, "docs", "research", "parity-scorecard.md"))
	if err != nil {
		t.Fatalf("read generated scorecard: %v", err)
	}
	scorecard := string(body)
	for _, phrase := range []string{
		"# goscrapling Parity Scorecard",
		"Text Extraction Speed Test",
		"Parser and selectors",
		"Static fetcher and response",
		"Spider runtime",
		"CLI shell and extract commands",
		"BenchmarkParserNestedText",
		"BenchmarkStaticFetcherLocalResponse",
		"BenchmarkSpiderSchedulerFingerprint",
		"BenchmarkCLIExtractFixture",
		"No live network, live browser, or live LLM",
	} {
		if !strings.Contains(scorecard, phrase) {
			t.Fatalf("scorecard missing %q:\n%s", phrase, scorecard)
		}
	}
}

func TestProgressMapWriteEndToEnd(t *testing.T) {
	root := t.TempDir()
	writeMinimalAppMapFixture(t, root)
	binary := buildProgressBinary(t)

	result := runProgressBinary(t, binary, "--repo-root", root, "map-write")
	if result.err != nil {
		t.Fatalf("progress map-write failed: %v\nstdout: %s\nstderr: %s", result.err, result.stdout, result.stderr)
	}
	if !strings.Contains(result.stdout, "app-map: regenerated docs/content/building-goscrapling/architecture_plan/upstream-app-map.md") {
		t.Fatalf("stdout missing map-write summary: %q", result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}

	body, err := os.ReadFile(filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "upstream-app-map.md"))
	if err != nil {
		t.Fatalf("read generated app map markdown: %v", err)
	}
	if got := string(body); !strings.Contains(got, "# Upstream Scrapling App Map") || !strings.Contains(got, "Parser core") {
		t.Fatalf("generated app map markdown missing expected content: %s", got)
	}
}

func TestProgressMapValidateErrorsExitOneEndToEnd(t *testing.T) {
	root := t.TempDir()
	writeMinimalAppMapFixture(t, root)
	appMapPath := filepath.Join(root, "docs", "content", "building-goscrapling", "architecture_plan", "upstream-app-map.json")
	body, err := os.ReadFile(appMapPath)
	if err != nil {
		t.Fatalf("read app map fixture: %v", err)
	}
	body = []byte(strings.Replace(string(body),
		`"translation_suitability": "manual_rewrite"`,
		`"translation_suitability": "manual_rewrite",
      "static_reference_paths": ["../outside.txt"]`,
		1,
	))
	if err := os.WriteFile(appMapPath, body, 0o644); err != nil {
		t.Fatalf("write app map fixture: %v", err)
	}

	binary := buildProgressBinary(t)
	result := runProgressBinary(t, binary, "--repo-root", root, "map-validate")
	if result.err == nil {
		t.Fatal("expected progress map-validate failure")
	}
	var exitErr *exec.ExitError
	if !errors.As(result.err, &exitErr) {
		t.Fatalf("error = %T %v, want *exec.ExitError", result.err, result.err)
	}
	if got := exitErr.ExitCode(); got != 1 {
		t.Fatalf("exit code = %d, want 1\nstdout: %s\nstderr: %s", got, result.stdout, result.stderr)
	}
	if !strings.Contains(result.stderr, "static reference path escapes repo root ../outside.txt") {
		t.Fatalf("stderr missing validation message: %q", result.stderr)
	}
}

type commandResult struct {
	stdout string
	stderr string
	err    error
}

func buildProgressBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "progress")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build progress binary: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return binary
}

func runProgressBinary(t *testing.T, binary string, args ...string) commandResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return commandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
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
