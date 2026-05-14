package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIExtractGet(t *testing.T) {
	t.Run("writes selected text from a local fixture response", func(t *testing.T) {
		var seenHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/products" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			seenHeader = r.Header.Get("X-Trace")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join("testdata", "products.html"))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "products.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{
			"extract", "get", server.URL + "/products", outputPath,
			"--css-selector", ".product",
			"-H", "X-Trace: cli-test",
			"--timeout", "2",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != "Trail Kit\nCamp Mug" {
			t.Fatalf("output text = %q", got)
		}
		if seenHeader != "cli-test" {
			t.Fatalf("expected request header X-Trace, got %q", seenHeader)
		}
		if !strings.Contains(stdout.String(), "wrote "+outputPath) {
			t.Fatalf("stdout missing output path: %q", stdout.String())
		}
	})

	t.Run("writes full HTML when the output extension is html", func(t *testing.T) {
		const html = `<html><body><h1>Full page</h1></body></html>`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(html))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "page.html")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{"extract", "get", server.URL, outputPath})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != html {
			t.Fatalf("output html = %q", got)
		}
	})

	t.Run("does not follow redirects when disabled", func(t *testing.T) {
		var hitFinal bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/redirect":
				w.Header().Set("Location", "/final")
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusFound)
				_, _ = w.Write([]byte(`<html><body><p class="notice">redirect held</p></body></html>`))
			case "/final":
				hitFinal = true
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(`<html><body><p class="notice">final page</p></body></html>`))
			default:
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "redirect.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{
			"extract", "get", server.URL + "/redirect", outputPath,
			"--css-selector", ".notice",
			"--no-follow-redirects",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != "redirect held" {
			t.Fatalf("output text = %q", got)
		}
		if hitFinal {
			t.Fatal("redirect final endpoint was hit despite --no-follow-redirects")
		}
	})

	t.Run("returns parse errors for malformed headers", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "broken.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{
			"extract", "get", "https://example.com", outputPath,
			"-H", "not-a-header",
		})

		if !errors.Is(err, ErrParse) {
			t.Fatalf("error = %v, want ErrParse", err)
		}
		if _, statErr := os.Stat(outputPath); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("output file exists after parse error: %v", statErr)
		}
	})
}

func TestCLIExtractMethods(t *testing.T) {
	t.Run("POST sends form body and query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("page"); got != "1" {
				t.Fatalf("query page = %q, want 1", got)
			}
			body := readRequestBody(t, r)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><div id="body">` + body + `</div></body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "post.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{
			"extract", "post", server.URL + "/submit", outputPath,
			"--data", "name=trail",
			"-p", "page=1",
			"-s", "#body",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != "name=trail" {
			t.Fatalf("output body = %q", got)
		}
	})

	t.Run("PUT sends JSON with content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", r.Method)
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("content-type = %q, want application/json", got)
			}
			body := readRequestBody(t, r)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><div id="json">` + body + `</div></body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "put.txt")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{
			"extract", "put", server.URL + "/resource", outputPath,
			"--json", `{"name":"mug"}`,
			"--css-selector", "#json",
		})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != `{"name":"mug"}` {
			t.Fatalf("output json = %q", got)
		}
	})

	t.Run("DELETE uses delete method and writes HTML", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("method = %s, want DELETE", r.Method)
			}
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><h1>Deleted</h1></body></html>`))
		}))
		defer server.Close()

		outputPath := filepath.Join(t.TempDir(), "delete.html")
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{"extract", "delete", server.URL + "/resource", outputPath})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		body, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if got := string(body); got != `<html><body><h1>Deleted</h1></body></html>` {
			t.Fatalf("output html = %q", got)
		}
	})
}

func readRequestBody(t *testing.T, r *http.Request) string {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	return string(body)
}
