package browser

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestStealthBrowserControls(t *testing.T) {
	t.Run("forwards explicit stealth controls without hidden defaults", func(t *testing.T) {
		engine := &recordingBrowserEngine{result: browserOptionsHTMLResult("https://example.com/rendered")}
		fetcher := BrowserFetcher{Engine: engine}

		_, err := fetcher.Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Proxy:      BrowserProxyOptions{Server: "http://proxy.example:8080", Username: "agent", Password: "secret"},
			RealChrome: true,
			ExtraFlags: []string{"--window-size=1200,900"},
			Stealth: BrowserStealthOptions{
				Enabled:         true,
				GenerateHeaders: true,
				GoogleReferer:   true,
				HideCanvas:      true,
				BlockWebRTC:     true,
				DisableWebGL:    true,
			},
		})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}

		request := engine.request
		if !request.Stealth.Enabled || !request.Stealth.GenerateHeaders || !request.Stealth.GoogleReferer || !request.Stealth.HideCanvas || !request.Stealth.BlockWebRTC || !request.Stealth.DisableWebGL {
			t.Fatalf("stealth controls were not forwarded: %#v", request.Stealth)
		}
		if request.Headers.Get("Referer") != "https://www.google.com/" {
			t.Fatalf("referer = %q", request.Headers.Get("Referer"))
		}
		if ua := request.Headers.Get("User-Agent"); !strings.Contains(ua, "Chrome/147") || request.UserAgent != ua {
			t.Fatalf("generated user agent = %q request.UserAgent=%q", ua, request.UserAgent)
		}
		if request.Headers.Get("Accept") == "" || request.Headers.Get("Accept-Language") == "" {
			t.Fatalf("generated headers missing Accept/Accept-Language: %#v", request.Headers)
		}
		wantFlags := []string{
			"--window-size=1200,900",
			"--force-webrtc-ip-handling-policy=disable_non_proxied_udp",
			"--disable-features=WebGL,WebGL2",
		}
		if !reflect.DeepEqual(request.ExtraFlags, wantFlags) {
			t.Fatalf("extra flags = %#v, want %#v", request.ExtraFlags, wantFlags)
		}
		if script := browserStealthInitScript(request); !strings.Contains(script, "HTMLCanvasElement") {
			t.Fatalf("canvas stealth script missing canvas hook: %q", script)
		}
	})

	t.Run("preserves caller headers unless a stealth option asks to replace them", func(t *testing.T) {
		engine := &recordingBrowserEngine{result: browserOptionsHTMLResult("https://example.com/rendered")}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Headers: http.Header{
				"User-Agent":      []string{"caller-agent/1.0"},
				"Referer":         []string{"https://caller.example/"},
				"Accept-Language": []string{"fr-FR"},
			},
			Stealth: BrowserStealthOptions{Enabled: true, GenerateHeaders: true},
		})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}
		if got := engine.request.Headers.Get("User-Agent"); got != "caller-agent/1.0" {
			t.Fatalf("User-Agent was overwritten: %q", got)
		}
		if got := engine.request.Headers.Get("Referer"); got != "https://caller.example/" {
			t.Fatalf("Referer was overwritten without GoogleReferer: %q", got)
		}
		if got := engine.request.Headers.Get("Accept-Language"); got != "fr-FR" {
			t.Fatalf("Accept-Language was overwritten: %q", got)
		}

		_, err = (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Headers: http.Header{"Referer": []string{"https://caller.example/"}},
			Stealth: BrowserStealthOptions{Enabled: true, GoogleReferer: true},
		})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}
		if got := engine.request.Headers.Get("Referer"); got != "https://www.google.com/" {
			t.Fatalf("GoogleReferer did not replace referer: %q", got)
		}
	})

	t.Run("rejects unsupported challenge solving claims before engine fetch", func(t *testing.T) {
		engine := &recordingBrowserEngine{result: browserOptionsHTMLResult("https://example.com/rendered")}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Stealth: BrowserStealthOptions{Enabled: true, SolveCloudflare: true},
		})
		if !errors.Is(err, ErrUnsupportedBrowserChallenge) || !errors.Is(err, ErrBrowserOptions) {
			t.Fatalf("Fetch error = %v, want ErrBrowserOptions and ErrUnsupportedBrowserChallenge", err)
		}
		if engine.request.URL != "" {
			t.Fatalf("engine was called for unsupported challenge solving: %#v", engine.request)
		}
	})
}
