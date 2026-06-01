package clitest

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Result captures a goscrapling CLI process invocation.
type Result struct {
	Stdout string
	Stderr string
	Err    error
}

// BuildBinary builds the cmd/goscrapling command for CLI integration tests.
func BuildBinary(t testing.TB) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "goscrapling")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	cmd.Dir = commandDir(t)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build goscrapling binary: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return binary
}

// RunBinary runs a previously built goscrapling test binary.
func RunBinary(t testing.TB, binary string, args ...string) Result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
}

func commandDir(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate clitest source file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
