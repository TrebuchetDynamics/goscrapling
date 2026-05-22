package fetchers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyRotator(t *testing.T) {
	t.Run("cycles string and dictionary-style proxy configuration", func(t *testing.T) {
		rotator, err := NewProxyRotator([]any{
			"http://proxy-one.test:8080",
			map[string]string{
				"server":   "http://proxy-two.test:8080",
				"username": "user",
				"password": "pass",
			},
		})
		if err != nil {
			t.Fatalf("NewProxyRotator: %v", err)
		}
		if got := rotator.Len(); got != 2 {
			t.Fatalf("rotator length = %d, want 2", got)
		}

		first, err := rotator.Next()
		if err != nil {
			t.Fatalf("first proxy: %v", err)
		}
		if first.URL != "http://proxy-one.test:8080" {
			t.Fatalf("first proxy URL = %q", first.URL)
		}

		second, err := rotator.Next()
		if err != nil {
			t.Fatalf("second proxy: %v", err)
		}
		if second.URL != "http://proxy-two.test:8080" {
			t.Fatalf("second proxy URL = %q", second.URL)
		}
		if second.Auth == nil || second.Auth.Username != "user" || second.Auth.Password != "pass" {
			t.Fatalf("dictionary auth = %#v", second.Auth)
		}

		third, err := rotator.Next()
		if err != nil {
			t.Fatalf("third proxy: %v", err)
		}
		if third.URL != first.URL {
			t.Fatalf("cyclic proxy URL = %q, want %q", third.URL, first.URL)
		}

		proxies := rotator.Proxies()
		proxies[0].URL = "http://mutated.invalid:8080"
		afterMutation, err := rotator.Next()
		if err != nil {
			t.Fatalf("proxy after mutation: %v", err)
		}
		if afterMutation.URL != second.URL {
			t.Fatalf("proxies copy mutation changed rotator state: %q", afterMutation.URL)
		}
	})

	t.Run("uses custom strategy hooks", func(t *testing.T) {
		rotator, err := NewProxyRotator(
			[]any{"http://proxy-one.test:8080", "http://proxy-two.test:8080"},
			WithProxyRotationStrategy(func(proxies []ProxyOptions, currentIndex int) (ProxyOptions, int) {
				return proxies[len(proxies)-1], currentIndex
			}),
		)
		if err != nil {
			t.Fatalf("NewProxyRotator: %v", err)
		}

		for i := 0; i < 3; i++ {
			proxy, err := rotator.Next()
			if err != nil {
				t.Fatalf("custom strategy proxy %d: %v", i, err)
			}
			if proxy.URL != "http://proxy-two.test:8080" {
				t.Fatalf("custom strategy proxy %d URL = %q", i, proxy.URL)
			}
		}
	})

	t.Run("rejects exhausted or invalid proxy configuration", func(t *testing.T) {
		if _, err := NewProxyRotator(nil); !errors.Is(err, ErrRequestOptions) {
			t.Fatalf("empty rotator error = %v, want ErrRequestOptions", err)
		}
		if _, err := NewProxyRotator([]any{map[string]string{"username": "user"}}); !errors.Is(err, ErrRequestOptions) {
			t.Fatalf("missing server error = %v, want ErrRequestOptions", err)
		}
		if _, err := NewProxyRotator([]any{42}); !errors.Is(err, ErrRequestOptions) {
			t.Fatalf("invalid proxy type error = %v, want ErrRequestOptions", err)
		}
	})

	t.Run("session uses the next rotated proxy for each request", func(t *testing.T) {
		firstProxy := proxyFixture(t, "first-rotated")
		secondProxy := proxyFixture(t, "second-rotated")
		rotator, err := NewProxyRotator([]any{firstProxy.URL, secondProxy.URL})
		if err != nil {
			t.Fatalf("NewProxyRotator: %v", err)
		}

		session, err := NewFetcherSession(FetcherSessionOptions{ProxyRotator: rotator})
		if err != nil {
			t.Fatalf("NewFetcherSession: %v", err)
		}
		first, err := session.Get("http://example.test/first", RequestOptions{Retries: 1})
		if err != nil {
			t.Fatalf("first session request: %v", err)
		}
		second, err := session.Get("http://example.test/second", RequestOptions{Retries: 1})
		if err != nil {
			t.Fatalf("second session request: %v", err)
		}

		if got := first.Meta()["proxy"]; got != firstProxy.URL {
			t.Fatalf("first proxy meta = %#v, want %q", got, firstProxy.URL)
		}
		if first.CSS(".first-rotated").Len() != 1 {
			t.Fatalf("first response did not come through first proxy: %q", first.Text())
		}
		if got := second.Meta()["proxy"]; got != secondProxy.URL {
			t.Fatalf("second proxy meta = %#v, want %q", got, secondProxy.URL)
		}
		if second.CSS(".second-rotated").Len() != 1 {
			t.Fatalf("second response did not come through second proxy: %q", second.Text())
		}
	})

	t.Run("proxy errors retry with the next rotated proxy", func(t *testing.T) {
		brokenProxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			closeRequestConnection(t, w)
		}))
		t.Cleanup(brokenProxy.Close)
		healthyProxy := proxyFixture(t, "healthy-rotated")

		rotator, err := NewProxyRotator([]any{brokenProxy.URL, healthyProxy.URL})
		if err != nil {
			t.Fatalf("NewProxyRotator: %v", err)
		}

		response, err := (Fetcher{}).Get("http://example.test/retry", RequestOptions{
			ProxyRotator: rotator,
			Retries:      2,
		})
		if err != nil {
			t.Fatalf("rotated retry returned error: %v", err)
		}
		if got := response.Meta()["proxy"]; got != healthyProxy.URL {
			t.Fatalf("retry proxy meta = %#v, want %q", got, healthyProxy.URL)
		}
		if response.CSS(".healthy-rotated").Len() != 1 {
			t.Fatalf("retry response did not come through healthy proxy: %q", response.Text())
		}
	})

	t.Run("rejects combining static proxy and proxy rotator", func(t *testing.T) {
		rotator, err := NewProxyRotator([]any{"http://rotated-proxy.test:8080"})
		if err != nil {
			t.Fatalf("NewProxyRotator: %v", err)
		}

		_, err = (Fetcher{}).Get("http://example.test/conflict", RequestOptions{
			Proxy:        ProxyOptions{URL: "http://static-proxy.test:8080"},
			ProxyRotator: rotator,
			Retries:      1,
		})
		if !errors.Is(err, ErrRequestOptions) {
			t.Fatalf("proxy conflict error = %v, want ErrRequestOptions", err)
		}
	})
}
