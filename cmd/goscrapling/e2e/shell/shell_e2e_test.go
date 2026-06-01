package shell_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/cmd/goscrapling/internal/clitest"
)

func TestGoscraplingShellBinaryEndToEnd(t *testing.T) {
	var seen []struct {
		method string
		path   string
		header string
		body   string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		seen = append(seen, struct {
			method string
			path   string
			header string
			body   string
		}{
			method: r.Method,
			path:   r.URL.Path,
			header: r.Header.Get("X-Shell-E2E"),
			body:   string(body),
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/products":
			fmt.Fprint(w, `<html><body><article class="product">Trail Kit</article><article class="product">Camp Mug</article></body></html>`)
		case "/submit":
			fmt.Fprintf(w, `<html><body><span id="method">%s</span><pre id="payload">%s</pre></body></html>`, r.Method, body)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	curlCommand := fmt.Sprintf(`curl -X POST %q -H 'X-Shell-E2E: curl' --data-raw '{"sku":"trail-kit"}'`, server.URL+"/submit")
	script := strings.Join([]string{
		fmt.Sprintf("get(%q)", server.URL+"/products"),
		`print(response.status)`,
		`print(len(page.css(".product")))`,
		`print(page.css(".product::text").text())`,
		fmt.Sprintf("curl2fetcher(%q)", curlCommand),
		`print(response.status)`,
		`print(page.css("#method::text").get(""))`,
		`print(page.css("#payload::text").get(""))`,
		`print(len(pages))`,
	}, "; ")

	binary := clitest.BuildBinary(t)
	result := clitest.RunBinary(t, binary, "shell", "-c", script)
	if result.Err != nil {
		t.Fatalf("goscrapling shell failed: %v\nstdout: %s\nstderr: %s", result.Err, result.Stdout, result.Stderr)
	}

	want := strings.Join([]string{
		"200",
		"2",
		"Trail Kit",
		"Camp Mug",
		"200",
		"POST",
		`{"sku":"trail-kit"}`,
		"2",
		"",
	}, "\n")
	if result.Stdout != want {
		t.Fatalf("stdout = %q, want %q", result.Stdout, want)
	}
	if result.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.Stderr)
	}
	if len(seen) != 2 {
		t.Fatalf("seen requests = %#v, want two", seen)
	}
	if seen[0].method != http.MethodGet || seen[0].path != "/products" || seen[0].header != "" {
		t.Fatalf("first request = %#v", seen[0])
	}
	if seen[1].method != http.MethodPost || seen[1].path != "/submit" || seen[1].header != "curl" || seen[1].body != `{"sku":"trail-kit"}` {
		t.Fatalf("second request = %#v", seen[1])
	}
}
