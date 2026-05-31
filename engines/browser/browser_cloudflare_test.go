package browser

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestCloudflareBoundary(t *testing.T) {
	t.Run("status helper documents unsupported disabled-by-default boundary", func(t *testing.T) {
		boundary := CloudflareChallengeBoundary()
		if boundary.Supported || boundary.DefaultEnabled {
			t.Fatalf("boundary = %#v, want unsupported and disabled by default", boundary)
		}
		if boundary.OptionName != "SolveCloudflare" || !errors.Is(boundary.Err, ErrUnsupportedBrowserChallenge) {
			t.Fatalf("boundary option/error = %#v / %v", boundary, boundary.Err)
		}
		for _, want := range []string{"unsupported", "deterministic local challenge fixture", "operator-visible controls"} {
			if !strings.Contains(boundary.Message, want) {
				t.Fatalf("boundary message %q missing %q", boundary.Message, want)
			}
		}
	})

	t.Run("normal and stealth controls do not solve challenges by default", func(t *testing.T) {
		engine := &cloudflareBoundaryEngine{}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/protected", BrowserOptions{
			Stealth: BrowserStealthOptions{Enabled: true, GenerateHeaders: true, GoogleReferer: true},
		})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}
		if engine.request.Stealth.SolveCloudflare {
			t.Fatalf("SolveCloudflare unexpectedly enabled: %#v", engine.request.Stealth)
		}
	})

	t.Run("solve cloudflare fails before browser engine work", func(t *testing.T) {
		engine := &cloudflareBoundaryEngine{}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/protected", BrowserOptions{
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

type cloudflareBoundaryEngine struct {
	request BrowserRequest
}

func (e *cloudflareBoundaryEngine) Fetch(_ context.Context, request BrowserRequest) (BrowserResult, error) {
	e.request = request
	return BrowserResult{
		URL:        request.URL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Body:       []byte("<html><body>ok</body></html>"),
	}, nil
}
