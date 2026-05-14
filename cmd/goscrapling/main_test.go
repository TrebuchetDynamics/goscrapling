package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGoscraplingExtractGetEndToEnd(t *testing.T) {
	var seenHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		seenHeader = r.Header.Get("X-E2E")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><article class="product">Trail <span>Kit</span></article><article class="product">Camp Mug</article></body></html>`))
	}))
	t.Cleanup(server.Close)

	binary := buildGoscraplingBinary(t)
	outputPath := filepath.Join(t.TempDir(), "products.txt")
	result := runGoscraplingBinary(t, binary,
		"extract", "get", server.URL+"/products", outputPath,
		"--css-selector", ".product",
		"-H", "X-E2E: binary",
		"--timeout", "2",
	)
	if result.err != nil {
		t.Fatalf("goscrapling extract failed: %v\nstdout: %s\nstderr: %s", result.err, result.stdout, result.stderr)
	}

	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got := string(body); got != "Trail Kit\nCamp Mug" {
		t.Fatalf("output text = %q", got)
	}
	if seenHeader != "binary" {
		t.Fatalf("X-E2E header = %q, want binary", seenHeader)
	}
	if !strings.Contains(result.stdout, "wrote "+outputPath) {
		t.Fatalf("stdout missing output path: %q", result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
}

func TestGoscraplingExtractPostJSONEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("query page = %q, want 2", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want application/json", got)
		}
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><pre id="payload">` + string(requestBody) + `</pre></body></html>`))
	}))
	t.Cleanup(server.Close)

	binary := buildGoscraplingBinary(t)
	outputPath := filepath.Join(t.TempDir(), "payload.txt")
	result := runGoscraplingBinary(t, binary,
		"extract", "post", server.URL+"/submit", outputPath,
		"--json", `{"name":"camp-mug"}`,
		"--params", "page=2",
		"--css-selector", "#payload",
	)
	if result.err != nil {
		t.Fatalf("goscrapling extract post failed: %v\nstdout: %s\nstderr: %s", result.err, result.stdout, result.stderr)
	}

	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got := string(body); got != `{"name":"camp-mug"}` {
		t.Fatalf("output text = %q", got)
	}
}

func TestGoscraplingParseErrorsExitTwoEndToEnd(t *testing.T) {
	binary := buildGoscraplingBinary(t)
	outputPath := filepath.Join(t.TempDir(), "broken.txt")
	result := runGoscraplingBinary(t, binary,
		"extract", "get", "https://example.com", outputPath,
		"-H", "not-a-header",
	)

	if result.err == nil {
		t.Fatal("expected goscrapling parse error")
	}
	var exitErr *exec.ExitError
	if !errors.As(result.err, &exitErr) {
		t.Fatalf("error = %T %v, want *exec.ExitError", result.err, result.err)
	}
	if got := exitErr.ExitCode(); got != 2 {
		t.Fatalf("exit code = %d, want 2\nstdout: %s\nstderr: %s", got, result.stdout, result.stderr)
	}
	if !strings.Contains(result.stderr, "parse error: headers must use") {
		t.Fatalf("stderr missing parse message: %q", result.stderr)
	}
	if _, err := os.Stat(outputPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output file exists after parse error: %v", err)
	}
}

type commandResult struct {
	stdout string
	stderr string
	err    error
}

func buildGoscraplingBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "goscrapling")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build goscrapling binary: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return binary
}

func runGoscraplingBinary(t *testing.T, binary string, args ...string) commandResult {
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
