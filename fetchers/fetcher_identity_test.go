package fetchers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestStaticFetcherIdentityOptions(t *testing.T) {
	t.Run("explicit stealthy headers add browser-like defaults without overwriting caller headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("User-Agent"); got != "Custom/1.0" {
				t.Fatalf("user-agent = %q, want caller header", got)
			}
			if got := r.Header.Get("Accept"); got != "application/json" {
				t.Fatalf("accept = %q, want caller header", got)
			}
			if got := r.Header.Get("Referer"); got != "https://www.google.com/" {
				t.Fatalf("referer = %q, want Google referer", got)
			}
			if got := r.Header.Get("Accept-Language"); got == "" {
				t.Fatal("accept-language was not generated")
			}
			if got := r.Header.Get("Upgrade-Insecure-Requests"); got != "1" {
				t.Fatalf("upgrade-insecure-requests = %q, want 1", got)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="identity">ok</article></body></html>`)
		}))
		t.Cleanup(server.Close)

		response, err := (Fetcher{}).Get(server.URL, RequestOptions{
			StealthyHeaders: Bool(true),
			Headers: http.Header{
				"User-Agent": []string{"Custom/1.0"},
				"Accept":     []string{"application/json"},
			},
			Retries: 1,
		})
		if err != nil {
			t.Fatalf("Fetcher.Get returned error: %v", err)
		}
		if response.CSS(".identity").Len() != 1 {
			t.Fatalf("identity response was not parsed: %q", response.Text())
		}
	})

	t.Run("explicitly disabled stealthy headers do not modify request metadata", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="plain">ok</article></body></html>`)
		}))
		t.Cleanup(server.Close)

		response, err := (Fetcher{}).Get(server.URL, RequestOptions{
			StealthyHeaders: Bool(false),
			Retries:         1,
		})
		if err != nil {
			t.Fatalf("Fetcher.Get returned error: %v", err)
		}
		headers := response.Request().Headers
		if got := headers.Get("Referer"); got != "" {
			t.Fatalf("referer metadata = %q, want empty", got)
		}
		if got := headers.Get("Upgrade-Insecure-Requests"); got != "" {
			t.Fatalf("upgrade-insecure-requests metadata = %q, want empty", got)
		}
	})

	t.Run("unsupported impersonation and HTTP3 fail before sending", func(t *testing.T) {
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(server.Close)

		_, err := (Fetcher{}).Get(server.URL, RequestOptions{
			Impersonate: "chrome",
			Retries:     1,
		})
		if !errors.Is(err, ErrRequestOptions) || !errors.Is(err, ErrUnsupportedStaticImpersonation) {
			t.Fatalf("impersonate error = %v, want ErrRequestOptions and ErrUnsupportedStaticImpersonation", err)
		}

		_, err = (Fetcher{}).Get(server.URL, RequestOptions{
			HTTP3:   Bool(true),
			Retries: 1,
		})
		if !errors.Is(err, ErrRequestOptions) || !errors.Is(err, ErrUnsupportedHTTP3) {
			t.Fatalf("HTTP3 error = %v, want ErrRequestOptions and ErrUnsupportedHTTP3", err)
		}

		if got := requests.Load(); got != 0 {
			t.Fatalf("server saw %d requests, want 0", got)
		}
	})
}
