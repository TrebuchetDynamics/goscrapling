package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
