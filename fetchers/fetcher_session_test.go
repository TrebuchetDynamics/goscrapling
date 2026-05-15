package fetchers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFetcherSessionOptionMergingAndCookies(t *testing.T) {
	var connections atomic.Int64

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/seed":
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: "abc123"})
		case "/echo":
			cookie, err := r.Cookie("sid")
			if err != nil {
				t.Errorf("expected persisted sid cookie: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Fprintf(
				w,
				`<html><body><article class="echo" data-session="%s" data-mode="%s" data-cookie="%s" data-body="%s"></article></body></html>`,
				r.Header.Get("X-Session"),
				r.Header.Get("X-Mode"),
				cookie.Value,
				string(body),
			)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)

	session, err := NewFetcherSession(FetcherSessionOptions{
		Headers: http.Header{
			"X-Session": []string{"session-default"},
			"X-Mode":    []string{"session-default"},
		},
	})
	if err != nil {
		t.Fatalf("NewFetcherSession returned error: %v", err)
	}

	seed, err := session.Get(server.URL+"/seed", RequestOptions{
		Headers: http.Header{
			"X-Mode": []string{"seed-request"},
		},
	})
	if err != nil {
		t.Fatalf("session.Get returned error: %v", err)
	}
	if got := seed.Request().Headers.Get("X-Session"); got != "session-default" {
		t.Fatalf("expected default header on seed request, got %q", got)
	}
	if got := seed.Request().Headers.Get("X-Mode"); got != "seed-request" {
		t.Fatalf("expected request header override on seed request, got %q", got)
	}

	echo, err := session.Post(server.URL+"/echo", RequestOptions{
		Headers: http.Header{
			"X-Mode": []string{"echo-request"},
		},
		Body: strings.NewReader("payload"),
	})
	if err != nil {
		t.Fatalf("session.Post returned error: %v", err)
	}

	request := echo.Request()
	if got := request.Headers.Get("X-Session"); got != "session-default" {
		t.Fatalf("expected default header on echo request, got %q", got)
	}
	if got := request.Headers.Get("X-Mode"); got != "echo-request" {
		t.Fatalf("expected request header override on echo request, got %q", got)
	}

	first, ok := echo.CSS(".echo").First()
	if !ok {
		t.Fatal("expected parsed echo article")
	}
	if got, ok := first.Attr("data-session"); !ok || got != "session-default" {
		t.Fatalf("expected merged session header, got %q ok=%v", got, ok)
	}
	if got, ok := first.Attr("data-mode"); !ok || got != "echo-request" {
		t.Fatalf("expected request override header, got %q ok=%v", got, ok)
	}
	if got, ok := first.Attr("data-cookie"); !ok || got != "abc123" {
		t.Fatalf("expected persisted cookie, got %q ok=%v", got, ok)
	}
	if got, ok := first.Attr("data-body"); !ok || got != "payload" {
		t.Fatalf("expected posted body, got %q ok=%v", got, ok)
	}
	if got := connections.Load(); got != 1 {
		t.Fatalf("expected session to reuse one connection, got %d", got)
	}
}
