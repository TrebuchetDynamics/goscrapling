package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type seenRequest struct {
	method      string
	path        string
	query       string
	contentType string
	header      string
	body        string
}

func TestGoscraplingFullLocalEndToEnd(t *testing.T) {
	var seen []seenRequest
	record := func(r *http.Request) string {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		seen = append(seen, seenRequest{
			method:      r.Method,
			path:        r.URL.Path,
			query:       r.URL.RawQuery,
			contentType: r.Header.Get("Content-Type"),
			header:      r.Header.Get("X-Full-E2E"),
			body:        string(body),
		})
		return string(body)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := record(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		switch r.URL.Path {
		case "/products":
			fmt.Fprint(w, `<html><body><article class="product"><h2>Trail Kit</h2></article><article class="product"><h2>Camp Mug</h2></article></body></html>`)
		case "/submit":
			if r.Method != http.MethodPost {
				t.Fatalf("/submit method = %s, want POST", r.Method)
			}
			fmt.Fprintf(w, `<html><body><pre id="payload">%s</pre></body></html>`, body)
		case "/replace":
			if r.Method != http.MethodPut {
				t.Fatalf("/replace method = %s, want PUT", r.Method)
			}
			fmt.Fprintf(w, `<html><body><pre id="payload">%s</pre></body></html>`, body)
		case "/delete":
			if r.Method != http.MethodDelete {
				t.Fatalf("/delete method = %s, want DELETE", r.Method)
			}
			fmt.Fprint(w, `<html><body><p class="status">Deleted</p></body></html>`)
		case "/redirect":
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/final":
			fmt.Fprint(w, `<html><body><p class="notice">followed redirect</p></body></html>`)
		case "/held-redirect":
			w.Header().Set("Location", "/final")
			w.WriteHeader(http.StatusFound)
			fmt.Fprint(w, `<html><body><p class="notice">held redirect</p></body></html>`)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	binary := buildGoscraplingBinary(t)
	outputDir := t.TempDir()

	cases := []struct {
		name       string
		args       []string
		outputName string
		want       string
	}{
		{
			name:       "get selected text",
			outputName: "products.txt",
			want:       "Trail Kit\nCamp Mug",
			args: []string{
				"extract", "get", server.URL + "/products", filepath.Join(outputDir, "products.txt"),
				"--css-selector", ".product h2::text",
				"-H", "X-Full-E2E: get",
				"--timeout", "2",
			},
		},
		{
			name:       "post json body and query",
			outputName: "post.txt",
			want:       `{"name":"trail-kit"}`,
			args: []string{
				"extract", "post", server.URL + "/submit", filepath.Join(outputDir, "post.txt"),
				"--json", `{"name":"trail-kit"}`,
				"--params", "page=2",
				"--css-selector", "#payload::text",
				"-H", "X-Full-E2E: post",
			},
		},
		{
			name:       "put form body",
			outputName: "put.txt",
			want:       "name=camp-mug",
			args: []string{
				"extract", "put", server.URL + "/replace", filepath.Join(outputDir, "put.txt"),
				"--data", "name=camp-mug",
				"--css-selector", "#payload::text",
				"-H", "X-Full-E2E: put",
			},
		},
		{
			name:       "delete html output",
			outputName: "delete.html",
			want:       `<p class="status">Deleted</p>`,
			args: []string{
				"extract", "delete", server.URL + "/delete", filepath.Join(outputDir, "delete.html"),
				"--css-selector", ".status",
				"-H", "X-Full-E2E: delete",
			},
		},
		{
			name:       "follow redirect",
			outputName: "redirect.txt",
			want:       "followed redirect",
			args: []string{
				"extract", "get", server.URL + "/redirect", filepath.Join(outputDir, "redirect.txt"),
				"--css-selector", ".notice::text",
				"-H", "X-Full-E2E: redirect",
			},
		},
		{
			name:       "hold redirect when disabled",
			outputName: "held-redirect.txt",
			want:       "held redirect",
			args: []string{
				"extract", "get", server.URL + "/held-redirect", filepath.Join(outputDir, "held-redirect.txt"),
				"--css-selector", ".notice::text",
				"--no-follow-redirects",
				"-H", "X-Full-E2E: held-redirect",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := runGoscraplingBinary(t, binary, tc.args...)
			if result.err != nil {
				t.Fatalf("goscrapling failed: %v\nstdout: %s\nstderr: %s", result.err, result.stdout, result.stderr)
			}
			outputPath := filepath.Join(outputDir, tc.outputName)
			body, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("read output %s: %v", outputPath, err)
			}
			if got := strings.TrimSpace(string(body)); got != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
			if !strings.Contains(result.stdout, "wrote "+outputPath) {
				t.Fatalf("stdout missing output path: %q", result.stdout)
			}
			if result.stderr != "" {
				t.Fatalf("stderr = %q, want empty", result.stderr)
			}
		})
	}

	wantSeen := map[string]seenRequest{
		"/products":      {method: http.MethodGet, path: "/products", header: "get"},
		"/submit":        {method: http.MethodPost, path: "/submit", query: "page=2", contentType: "application/json", header: "post", body: `{"name":"trail-kit"}`},
		"/replace":       {method: http.MethodPut, path: "/replace", contentType: "application/x-www-form-urlencoded", header: "put", body: "name=camp-mug"},
		"/delete":        {method: http.MethodDelete, path: "/delete", header: "delete"},
		"/redirect":      {method: http.MethodGet, path: "/redirect", header: "redirect"},
		"/final":         {method: http.MethodGet, path: "/final", header: "redirect"},
		"/held-redirect": {method: http.MethodGet, path: "/held-redirect", header: "held-redirect"},
	}
	for path, want := range wantSeen {
		got, ok := seenRequestByPath(seen, path)
		if !ok {
			t.Fatalf("missing request for %s; saw %#v", path, seen)
		}
		if got.method != want.method || got.query != want.query || got.contentType != want.contentType || got.header != want.header || got.body != want.body {
			t.Fatalf("request %s = %#v, want %#v", path, got, want)
		}
	}
}

func seenRequestByPath(seen []seenRequest, path string) (seenRequest, bool) {
	for _, request := range seen {
		if request.path == path {
			return request, true
		}
	}
	return seenRequest{}, false
}
