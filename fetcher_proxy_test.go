package goscrapling

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestStaticFetcherProxySupport(t *testing.T) {
	t.Run("routes requests through an explicit proxy with proxy auth", func(t *testing.T) {
		var originRequests atomic.Int64
		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			originRequests.Add(1)
			w.WriteHeader(http.StatusTeapot)
		}))
		t.Cleanup(origin.Close)

		proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != origin.URL+"/catalog" {
				t.Fatalf("proxy target URL = %q, want %q", r.URL.String(), origin.URL+"/catalog")
			}
			wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("proxy-user:proxy-pass"))
			if got := r.Header.Get("Proxy-Authorization"); got != wantAuth {
				t.Fatalf("proxy authorization = %q, want %q", got, wantAuth)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body><article class="proxied">via proxy</article></body></html>`)
		}))
		t.Cleanup(proxy.Close)

		response, err := (Fetcher{}).Get(origin.URL+"/catalog", RequestOptions{
			Proxy: ProxyOptions{
				URL:  proxy.URL,
				Auth: &BasicAuth{Username: "proxy-user", Password: "proxy-pass"},
			},
			Retries: 1,
		})
		if err != nil {
			t.Fatalf("Fetcher.Get returned error: %v", err)
		}
		if got := originRequests.Load(); got != 0 {
			t.Fatalf("origin saw %d direct requests, want 0", got)
		}
		if got := response.Meta()["proxy"]; got != proxy.URL {
			t.Fatalf("response proxy meta = %#v, want %q", got, proxy.URL)
		}
		if proxied := response.CSS(".proxied"); proxied.Len() != 1 {
			t.Fatalf("expected proxied response to parse, got %d matches", proxied.Len())
		}
	})

	t.Run("uses session proxy defaults and allows per request proxy override", func(t *testing.T) {
		defaultProxy := proxyFixture(t, "default")
		overrideProxy := proxyFixture(t, "override")

		session, err := NewFetcherSession(FetcherSessionOptions{
			Proxy: ProxyOptions{URL: defaultProxy.URL},
		})
		if err != nil {
			t.Fatalf("NewFetcherSession: %v", err)
		}

		first, err := session.Get("http://example.test/default", RequestOptions{Retries: 1})
		if err != nil {
			t.Fatalf("session default proxy request returned error: %v", err)
		}
		if got := first.Meta()["proxy"]; got != defaultProxy.URL {
			t.Fatalf("default proxy meta = %#v, want %q", got, defaultProxy.URL)
		}
		if first.CSS(".default").Len() != 1 {
			t.Fatalf("expected default proxy response, got %q", first.Text())
		}

		second, err := session.Get("http://example.test/override", RequestOptions{
			Proxy:   ProxyOptions{URL: overrideProxy.URL},
			Retries: 1,
		})
		if err != nil {
			t.Fatalf("session override proxy request returned error: %v", err)
		}
		if got := second.Meta()["proxy"]; got != overrideProxy.URL {
			t.Fatalf("override proxy meta = %#v, want %q", got, overrideProxy.URL)
		}
		if second.CSS(".override").Len() != 1 {
			t.Fatalf("expected override proxy response, got %q", second.Text())
		}
	})

	t.Run("selects scheme-specific proxies", func(t *testing.T) {
		httpProxy := proxyFixture(t, "http-proxy")

		response, err := (Fetcher{}).Get("http://example.test/scheme", RequestOptions{
			Proxy: ProxyOptions{
				URLs: map[string]string{"http": httpProxy.URL},
			},
			Retries: 1,
		})
		if err != nil {
			t.Fatalf("Fetcher.Get with scheme proxy returned error: %v", err)
		}
		if got := response.Meta()["proxy"]; got != httpProxy.URL {
			t.Fatalf("scheme proxy meta = %#v, want %q", got, httpProxy.URL)
		}
		if response.CSS(".http-proxy").Len() != 1 {
			t.Fatalf("expected scheme proxy response, got %q", response.Text())
		}
	})

	t.Run("classifies proxy transport errors", func(t *testing.T) {
		brokenProxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			closeRequestConnection(t, w)
		}))
		t.Cleanup(brokenProxy.Close)

		_, err := (Fetcher{}).Get("http://example.test/broken", RequestOptions{
			Proxy:   ProxyOptions{URL: brokenProxy.URL},
			Retries: 1,
		})
		if !errors.Is(err, ErrProxyRequest) {
			t.Fatalf("expected ErrProxyRequest, got %v", err)
		}
		var fetchErr *FetcherError
		if !errors.As(err, &fetchErr) {
			t.Fatalf("expected FetcherError, got %T", err)
		}
		if fetchErr.Kind != FetcherErrorProxy {
			t.Fatalf("fetch error kind = %q, want %q", fetchErr.Kind, FetcherErrorProxy)
		}
	})

	t.Run("rejects invalid proxy options before sending", func(t *testing.T) {
		var requests atomic.Int64
		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(origin.Close)

		_, err := (Fetcher{}).Get(origin.URL, RequestOptions{
			Proxy: ProxyOptions{
				URL:  "http://proxy.invalid:8080",
				URLs: map[string]string{"http": "http://proxy.invalid:8081"},
			},
			Retries: 1,
		})
		if !errors.Is(err, ErrRequestOptions) {
			t.Fatalf("expected request options error, got %v", err)
		}
		if got := requests.Load(); got != 0 {
			t.Fatalf("origin saw %d requests, want 0", got)
		}
	})
}

func proxyFixture(t *testing.T, className string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "" || r.URL.Host == "" {
			t.Fatalf("proxy saw non-absolute target URL %q", r.URL.String())
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html><body><article class="%s">%s</article></body></html>`, className, strings.ReplaceAll(className, "-", " "))
	}))
	t.Cleanup(server.Close)
	return server
}
