package goscrapling

import (
	"encoding/json"
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

func TestStaticFetcherRequestOptions(t *testing.T) {
	t.Run("sends params data auth cookies and exposes response cookies", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			query := r.URL.Query()
			if got := query["existing"]; len(got) != 1 || got[0] != "1" {
				t.Fatalf("existing query = %#v, want [1]", got)
			}
			if got := query["page"]; len(got) != 1 || got[0] != "2" {
				t.Fatalf("page query = %#v, want [2]", got)
			}
			if got := query["tag"]; len(got) != 2 || got[0] != "go" || got[1] != "scrapling" {
				t.Fatalf("tag query = %#v, want [go scrapling]", got)
			}
			username, password, ok := r.BasicAuth()
			if !ok || username != "scrapling" || password != "secret" {
				t.Fatalf("basic auth = %q %q ok=%v, want scrapling secret true", username, password, ok)
			}
			if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Fatalf("content-type = %q, want form encoded", got)
			}
			sid, err := r.Cookie("sid")
			if err != nil || sid.Value != "cookie-wins" {
				t.Fatalf("sid cookie = %#v err=%v, want cookie-wins", sid, err)
			}
			mode, err := r.Cookie("mode")
			if err != nil || mode.Value != "map-value" {
				t.Fatalf("mode cookie = %#v err=%v, want map-value", mode, err)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if got := string(body); got != "key=value&multi=one&multi=two" {
				t.Fatalf("body = %q, want encoded data", got)
			}
			http.SetCookie(w, &http.Cookie{Name: "seen", Value: "yes"})
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="ok">request options</article></body></html>`)
		}))
		t.Cleanup(server.Close)

		response, err := (Fetcher{}).Post(server.URL+"/submit?existing=1", RequestOptions{
			Params: url.Values{
				"page": []string{"2"},
				"tag":  []string{"go", "scrapling"},
			},
			Data: url.Values{
				"key":   []string{"value"},
				"multi": []string{"one", "two"},
			},
			Auth: &BasicAuth{Username: "scrapling", Password: "secret"},
			CookieValues: map[string]string{
				"sid":  "map-value",
				"mode": "map-value",
			},
			Cookies: []*http.Cookie{
				{Name: "sid", Value: "cookie-wins"},
			},
		})
		if err != nil {
			t.Fatalf("Fetcher.Post returned error: %v", err)
		}
		if got := response.Request().URL; !strings.Contains(got, "existing=1") || !strings.Contains(got, "page=2") {
			t.Fatalf("request URL = %q, want merged query params", got)
		}
		cookies := response.Cookies()
		if len(cookies) != 1 || cookies[0].Name != "seen" || cookies[0].Value != "yes" {
			t.Fatalf("response cookies = %#v, want seen=yes", cookies)
		}
	})

	t.Run("marshals JSON and respects explicit content type and authorization", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Content-Type"); got != "application/vnd.scrapling+json" {
				t.Fatalf("content-type = %q, want explicit vendor JSON", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer explicit" {
				t.Fatalf("authorization = %q, want explicit header", got)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode JSON body: %v", err)
			}
			if payload["key"] != "value" {
				t.Fatalf("payload = %#v, want key=value", payload)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"ok":true}`)
		}))
		t.Cleanup(server.Close)

		response, err := (Fetcher{}).Put(server.URL+"/json", RequestOptions{
			Headers: http.Header{
				"Content-Type":  []string{"application/vnd.scrapling+json"},
				"Authorization": []string{"Bearer explicit"},
			},
			JSON: map[string]string{"key": "value"},
			Auth: &BasicAuth{Username: "ignored", Password: "ignored"},
		})
		if err != nil {
			t.Fatalf("Fetcher.Put returned error: %v", err)
		}
		if got := response.Headers().Get("Content-Type"); got != "application/json" {
			t.Fatalf("response content-type = %q, want application/json", got)
		}
	})

	t.Run("marshals JSON with a default content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("content-type = %q, want application/json", got)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode JSON body: %v", err)
			}
			if payload["key"] != "value" {
				t.Fatalf("payload = %#v, want key=value", payload)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(server.Close)

		_, err := (Fetcher{}).Post(server.URL+"/json", RequestOptions{
			JSON: map[string]string{"key": "value"},
		})
		if err != nil {
			t.Fatalf("Fetcher.Post returned error: %v", err)
		}
	})

	t.Run("rejects ambiguous body options before sending", func(t *testing.T) {
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(server.Close)

		conflicts := []struct {
			name string
			opts RequestOptions
		}{
			{
				name: "body and json",
				opts: RequestOptions{
					Body: strings.NewReader("raw"),
					JSON: map[string]string{"key": "value"},
				},
			},
			{
				name: "body and data",
				opts: RequestOptions{
					Body: strings.NewReader("raw"),
					Data: url.Values{"key": []string{"value"}},
				},
			},
			{
				name: "data and json",
				opts: RequestOptions{
					Data: url.Values{"key": []string{"value"}},
					JSON: map[string]string{"key": "value"},
				},
			},
		}
		for _, tt := range conflicts {
			_, err := (Fetcher{}).Post(server.URL, tt.opts)
			if err == nil {
				t.Fatalf("%s: expected body option conflict to return an error", tt.name)
			}
		}
		if got := requests.Load(); got != 0 {
			t.Fatalf("server saw %d requests, want 0", got)
		}
	})

	t.Run("returns JSON marshal errors before sending", func(t *testing.T) {
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(server.Close)

		_, err := (Fetcher{}).Post(server.URL, RequestOptions{
			JSON: func() {},
		})
		if err == nil {
			t.Fatal("expected invalid JSON payload to return an error")
		}
		if got := requests.Load(); got != 0 {
			t.Fatalf("server saw %d requests, want 0", got)
		}
	})

	t.Run("verify false allows explicit self-signed TLS fetches", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="tls">ok</article></body></html>`)
		}))
		t.Cleanup(server.Close)

		_, err := (Fetcher{}).Get(server.URL, RequestOptions{Verify: Bool(false), Retries: 1})
		if err != nil {
			t.Fatalf("Fetcher.Get with verify=false returned error: %v", err)
		}
	})
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
