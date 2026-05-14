package goscrapling

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStaticFetcherMethods(t *testing.T) {
	statusByMethod := map[string]int{
		http.MethodGet:    http.StatusOK,
		http.MethodPost:   http.StatusCreated,
		http.MethodPut:    http.StatusAccepted,
		http.MethodDelete: http.StatusPartialContent,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status, ok := statusByMethod[r.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Request-Method", r.Method)
		w.WriteHeader(status)
		fmt.Fprintf(
			w,
			`<html><body><article class="request" data-method="%s" data-header="%s" data-body="%s">%s %s</article></body></html>`,
			r.Method,
			r.Header.Get("X-Client"),
			string(body),
			r.Method,
			string(body),
		)
	}))
	t.Cleanup(server.Close)

	fetcher := Fetcher{}
	tests := []struct {
		name   string
		method string
		path   string
		header string
		body   string
		call   func(string, RequestOptions) (*Response, error)
	}{
		{
			name:   "get",
			method: http.MethodGet,
			path:   "/get",
			header: "read-client",
			call:   fetcher.Get,
		},
		{
			name:   "post",
			method: http.MethodPost,
			path:   "/post",
			header: "create-client",
			body:   "create-product",
			call:   fetcher.Post,
		},
		{
			name:   "put",
			method: http.MethodPut,
			path:   "/put",
			header: "update-client",
			body:   "update-product",
			call:   fetcher.Put,
		},
		{
			name:   "delete",
			method: http.MethodDelete,
			path:   "/delete",
			header: "delete-client",
			call:   fetcher.Delete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := tt.call(server.URL+tt.path, RequestOptions{
				Headers: http.Header{
					"X-Client": []string{tt.header},
				},
				Body: strings.NewReader(tt.body),
			})
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.method, err)
			}

			if got := response.StatusCode(); got != statusByMethod[tt.method] {
				t.Fatalf("expected status %d, got %d", statusByMethod[tt.method], got)
			}
			if got := response.Headers().Get("X-Request-Method"); got != tt.method {
				t.Fatalf("expected response header method %q, got %q", tt.method, got)
			}

			request := response.Request()
			if request.Method != tt.method {
				t.Fatalf("expected request method %q, got %q", tt.method, request.Method)
			}
			if request.URL != server.URL+tt.path {
				t.Fatalf("expected request URL %q, got %q", server.URL+tt.path, request.URL)
			}
			if got := request.Headers.Get("X-Client"); got != tt.header {
				t.Fatalf("expected request X-Client %q, got %q", tt.header, got)
			}

			first, ok := response.CSS(".request").First()
			if !ok {
				t.Fatal("expected parsed request article")
			}
			if got, ok := first.Attr("data-method"); !ok || got != tt.method {
				t.Fatalf("expected parsed method %q, got %q ok=%v", tt.method, got, ok)
			}
			if got, ok := first.Attr("data-header"); !ok || got != tt.header {
				t.Fatalf("expected parsed header %q, got %q ok=%v", tt.header, got, ok)
			}
			if got, ok := first.Attr("data-body"); !ok || got != tt.body {
				t.Fatalf("expected parsed body %q, got %q ok=%v", tt.body, got, ok)
			}
		})
	}
}

func TestFetcherRedirectTimeoutRetryErrors(t *testing.T) {
	var flakyAttempts atomic.Int64
	privateRedirectBase := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/safe":
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/final":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="final">safe redirect</article></body></html>`)
		case "/disabled":
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/private-redirect":
			http.Redirect(w, r, privateRedirectBase+"/private", http.StatusFound)
		case "/slow":
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/flaky":
			if flakyAttempts.Add(1) < 3 {
				closeRequestConnection(t, w)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<html><body><article class="retry" data-attempts="%d"></article></body></html>`, flakyAttempts.Load())
		case "/always-close":
			closeRequestConnection(t, w)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	_, port, err := net.SplitHostPort(serverURL.Host)
	if err != nil {
		t.Fatalf("split server host: %v", err)
	}
	privateRedirectBase = "http://127.0.0.1:" + port
	localhostBase := "http://localhost:" + port

	fetcher := Fetcher{}

	safe, err := fetcher.Get(server.URL+"/safe", RequestOptions{})
	if err != nil {
		t.Fatalf("safe redirect returned error: %v", err)
	}
	if safe.URL() != server.URL+"/final" {
		t.Fatalf("expected final redirect URL, got %q", safe.URL())
	}
	if final := safe.CSS(".final"); final.Len() != 1 {
		t.Fatalf("expected final redirect body to be parsed, got %d matches", final.Len())
	}

	disabled, err := fetcher.Get(server.URL+"/disabled", RequestOptions{FollowRedirects: RedirectPolicyNone})
	if err != nil {
		t.Fatalf("disabled redirect returned error: %v", err)
	}
	if got := disabled.StatusCode(); got != http.StatusFound {
		t.Fatalf("expected disabled redirect to return 302 response, got %d", got)
	}
	if got := disabled.Headers().Get("Location"); got != "/final" {
		t.Fatalf("expected disabled redirect Location /final, got %q", got)
	}

	_, err = fetcher.Get(localhostBase+"/private-redirect", RequestOptions{})
	if !errors.Is(err, ErrPrivateAddressRedirect) {
		t.Fatalf("expected private redirect error, got %v", err)
	}

	_, err = fetcher.Get(server.URL+"/slow", RequestOptions{
		Timeout: 5 * time.Millisecond,
		Retries: 1,
	})
	if !errors.Is(err, ErrRequestTimeout) {
		t.Fatalf("expected timeout error, got %v", err)
	}

	retried, err := fetcher.Get(server.URL+"/flaky", RequestOptions{Retries: 3})
	if err != nil {
		t.Fatalf("retrying flaky endpoint returned error: %v", err)
	}
	first, ok := retried.CSS(".retry").First()
	if !ok {
		t.Fatal("expected retry article")
	}
	if got, ok := first.Attr("data-attempts"); !ok || got != "3" {
		t.Fatalf("expected three retry attempts, got %q ok=%v", got, ok)
	}

	_, err = fetcher.Get(server.URL+"/always-close", RequestOptions{Retries: 2})
	if !errors.Is(err, ErrRetryExhausted) {
		t.Fatalf("expected retry exhausted error, got %v", err)
	}
	var fetchErr *FetcherError
	if !errors.As(err, &fetchErr) {
		t.Fatalf("expected FetcherError, got %T", err)
	}
	if fetchErr.Attempts != 2 {
		t.Fatalf("expected 2 exhausted attempts, got %d", fetchErr.Attempts)
	}
}

func closeRequestConnection(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		t.Error("response writer does not support hijacking")
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		t.Errorf("hijack connection: %v", err)
		return
	}
	if err := conn.Close(); err != nil {
		t.Errorf("close hijacked connection: %v", err)
	}
}
