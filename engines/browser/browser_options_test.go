package browser

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestBrowserContextOptions(t *testing.T) {
	t.Run("passes explicit context options without hidden stealth", func(t *testing.T) {
		engine := &recordingBrowserEngine{result: browserOptionsHTMLResult("https://example.com/rendered")}
		fetcher := BrowserFetcher{Engine: engine}
		cookies := []*http.Cookie{{Name: "sid", Value: "abc123", Domain: "example.com", Path: "/"}}
		extraFlags := []string{"--disable-gpu", "--window-size=1200,900"}
		blockedDomains := []string{"ads.example.com", "tracker.example"}

		response, err := fetcher.Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Headers:          http.Header{"X-Trace": []string{"ctx-options"}},
			UserAgent:        "goscrapling-test/1.0",
			Cookies:          cookies,
			Locale:           "en-GB",
			TimezoneID:       "America/Monterrey",
			Proxy:            BrowserProxyOptions{Server: "http://proxy.example:8080", Username: "juan", Password: "secret"},
			CDPURL:           "ws://127.0.0.1:9222/devtools/browser/test",
			RealChrome:       true,
			DisableResources: true,
			BlockedDomains:   blockedDomains,
			BlockAds:         true,
			DNSOverHTTPS:     true,
			ExtraFlags:       extraFlags,
		})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}
		if response.StatusCode() != http.StatusOK {
			t.Fatalf("response status = %d", response.StatusCode())
		}

		request := engine.request
		if got := request.Headers.Get("X-Trace"); got != "ctx-options" {
			t.Fatalf("X-Trace header = %q", got)
		}
		if got := request.Headers.Get("User-Agent"); got != "goscrapling-test/1.0" {
			t.Fatalf("User-Agent header = %q", got)
		}
		if got := request.Headers.Get("Accept-Language"); got != "en-GB" {
			t.Fatalf("Accept-Language header = %q", got)
		}
		if request.UserAgent != "goscrapling-test/1.0" || request.Locale != "en-GB" || request.TimezoneID != "America/Monterrey" {
			t.Fatalf("identity context = useragent %q locale %q timezone %q", request.UserAgent, request.Locale, request.TimezoneID)
		}
		if !reflect.DeepEqual(request.Cookies, cookies) {
			t.Fatalf("cookies = %#v, want %#v", request.Cookies, cookies)
		}
		if request.Proxy.Server != "http://proxy.example:8080" || request.Proxy.Username != "juan" || request.Proxy.Password != "secret" {
			t.Fatalf("proxy = %#v", request.Proxy)
		}
		if request.CDPURL == "" || !request.RealChrome || !request.BlockAds || !request.DNSOverHTTPS || !request.DisableResources {
			t.Fatalf("browser context flags = %#v", request)
		}
		if !reflect.DeepEqual(request.BlockedDomains, blockedDomains) {
			t.Fatalf("blocked domains = %#v", request.BlockedDomains)
		}
		if !reflect.DeepEqual(request.ExtraFlags, extraFlags) {
			t.Fatalf("extra flags = %#v", request.ExtraFlags)
		}

		cookies[0].Value = "mutated"
		blockedDomains[0] = "mutated.example"
		extraFlags[0] = "--mutated"
		if request.Cookies[0].Value != "abc123" || request.BlockedDomains[0] != "ads.example.com" || request.ExtraFlags[0] != "--disable-gpu" {
			t.Fatalf("request did not defensively copy browser options: %#v", request)
		}
	})

	t.Run("validates unsupported proxy shapes before engine fetch", func(t *testing.T) {
		engine := &recordingBrowserEngine{result: browserOptionsHTMLResult("https://example.com/rendered")}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Proxy: BrowserProxyOptions{Server: "ftp://proxy.example:21"},
		})
		if !errors.Is(err, ErrBrowserOptions) {
			t.Fatalf("Fetch error = %v, want ErrBrowserOptions", err)
		}
		if engine.request.URL != "" {
			t.Fatalf("engine was called for invalid proxy: %#v", engine.request)
		}
	})

	t.Run("merges session and request context options", func(t *testing.T) {
		engine := &recordingBrowserEngine{result: browserOptionsHTMLResult("https://example.com/rendered")}
		session, err := NewBrowserSession(BrowserSessionOptions{
			Engine: engine,
			Options: BrowserOptions{
				UserAgent:      "session-agent/1.0",
				Cookies:        []*http.Cookie{{Name: "session", Value: "cookie"}},
				BlockedDomains: []string{"session-block.example"},
				BlockAds:       true,
			},
		})
		if err != nil {
			t.Fatalf("NewBrowserSession returned error: %v", err)
		}
		t.Cleanup(func() { _ = session.Close() })

		_, err = session.Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Locale:         "es-MX",
			Cookies:        []*http.Cookie{{Name: "request", Value: "cookie"}},
			BlockedDomains: []string{"request-block.example"},
		})
		if err != nil {
			t.Fatalf("session Fetch returned error: %v", err)
		}
		request := engine.request
		if request.UserAgent != "session-agent/1.0" || request.Headers.Get("User-Agent") != "session-agent/1.0" {
			t.Fatalf("session useragent was not inherited: %#v", request)
		}
		if request.Locale != "es-MX" || request.Headers.Get("Accept-Language") != "es-MX" {
			t.Fatalf("request locale was not applied: %#v", request)
		}
		if got := cookieNames(request.Cookies); !reflect.DeepEqual(got, []string{"session", "request"}) {
			t.Fatalf("cookies = %#v, want session then request", got)
		}
		if !reflect.DeepEqual(request.BlockedDomains, []string{"session-block.example", "request-block.example"}) {
			t.Fatalf("merged blocked domains = %#v", request.BlockedDomains)
		}
		if !request.BlockAds {
			t.Fatalf("session BlockAds was not inherited: %#v", request)
		}
	})
}

func browserOptionsHTMLResult(rawURL string) BrowserResult {
	return BrowserResult{
		URL:        rawURL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Body:       []byte(`<html><body><article class="ok">ok</article></body></html>`),
	}
}

func cookieNames(cookies []*http.Cookie) []string {
	names := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}
	return names
}
