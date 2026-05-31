package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIShellCurlHelpers(t *testing.T) {
	var seenMethod, seenTrace, seenCookie, seenBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenTrace = r.Header.Get("X-Trace")
		seenCookie = r.Header.Get("Cookie")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		seenBody = string(body)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<html><body><span id="method">%s</span><span id="query">%s</span></body></html>`, r.Method, r.URL.Query().Get("page"))
	}))
	defer server.Close()

	curlCommand := fmt.Sprintf(`curl -X POST %q -H 'X-Trace: shell' -H 'Cookie: session=abc; theme=dark' --data-raw '{"q":"kit"}'`, server.URL+"/submit?page=2")
	script := strings.Join([]string{
		fmt.Sprintf("print(uncurl(%q).method)", curlCommand),
		fmt.Sprintf("print(uncurl(%q).url)", curlCommand),
		fmt.Sprintf("print(uncurl(%q).header(\"X-Trace\"))", curlCommand),
		fmt.Sprintf("print(uncurl(%q).cookie(\"session\"))", curlCommand),
		fmt.Sprintf("print(uncurl(%q).param(\"page\"))", curlCommand),
		fmt.Sprintf("print(uncurl(%q).body)", curlCommand),
		fmt.Sprintf("curl2fetcher(%q)", curlCommand),
		`print(response.status)`,
		`print(page.css("#method::text").get(""))`,
		`print(page.css("#query::text").get(""))`,
	}, "; ")

	var stdout, stderr bytes.Buffer
	err := Run(&stdout, &stderr, []string{"shell", "-c", script})
	if err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
	}

	want := strings.Join([]string{
		"post",
		server.URL + "/submit",
		"shell",
		"abc",
		"2",
		`{"q":"kit"}`,
		"200",
		"POST",
		"2",
		"",
	}, "\n")
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if seenMethod != http.MethodPost || seenTrace != "shell" || !strings.Contains(seenCookie, "session=abc") || seenBody != `{"q":"kit"}` {
		t.Fatalf("request method=%q trace=%q cookie=%q body=%q", seenMethod, seenTrace, seenCookie, seenBody)
	}

	stdout.Reset()
	stderr.Reset()
	err = Run(&stdout, &stderr, []string{"shell", "-c", `print(uncurl("curl https://example.com --unsupported").method)`})
	if err == nil || !strings.Contains(err.Error(), "unsupported curl option") {
		t.Fatalf("unsupported curl error = %v", err)
	}
}

func TestCLIShell(t *testing.T) {
	t.Run("scripted shell updates page shortcuts and evaluates selectors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			switch r.URL.Path {
			case "/products":
				_, _ = w.Write([]byte(`<html><body><div class="product">Trail Kit</div><div class="product">Camp Mug</div></body></html>`))
			case "/detail":
				_, _ = w.Write([]byte(`<html><body><h1>Detail Page</h1></body></html>`))
			default:
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
		}))
		defer server.Close()

		script := strings.Join([]string{
			fmt.Sprintf("get(%q)", server.URL+"/products"),
			"print(response.status)",
			`print(len(page.css(".product")))`,
			fmt.Sprintf("get(%q)", server.URL+"/detail"),
			"print(page.url)",
			"print(len(pages))",
			`print(page.css("h1::text").get(""))`,
		}, "; ")

		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{"shell", "-c", script})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		want := strings.Join([]string{
			"200",
			"2",
			server.URL + "/detail",
			"2",
			"Detail Page",
			"",
		}, "\n")
		if got := stdout.String(); got != want {
			t.Fatalf("stdout = %q, want %q", got, want)
		}
	})
}
