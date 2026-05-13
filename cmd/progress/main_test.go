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
	if !strings.Contains(string(agentQueue), "Response metadata and selector contract") {
		t.Fatalf("agent queue missing generated row: %s", agentQueue)
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
