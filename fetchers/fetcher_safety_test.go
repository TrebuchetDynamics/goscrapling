package fetchers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFetcherSafetyControls(t *testing.T) {
	t.Run("obeys robots allow and disallow decisions from local fixtures", func(t *testing.T) {
		var robotsRequests atomic.Int64
		var pageRequests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/robots.txt":
				robotsRequests.Add(1)
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprint(w, "User-agent: *\nDisallow: /private\nAllow: /private/public\n")
			case "/private/secret":
				pageRequests.Add(1)
				fmt.Fprint(w, `<html><body>secret</body></html>`)
			case "/private/public/page":
				pageRequests.Add(1)
				fmt.Fprint(w, `<html><body><article class="ok">public</article></body></html>`)
			default:
				http.NotFound(w, r)
			}
		}))
		t.Cleanup(server.Close)

		fetcher := Fetcher{}
		_, err := fetcher.Get(server.URL+"/private/secret", RequestOptions{
			Safety: FetchSafetyOptions{ObeyRobots: true},
		})
		if !errors.Is(err, ErrRobotsBlocked) {
			t.Fatalf("expected robots blocked error, got %v", err)
		}
		if got := pageRequests.Load(); got != 0 {
			t.Fatalf("blocked path reached server %d times, want 0", got)
		}

		response, err := fetcher.Get(server.URL+"/private/public/page", RequestOptions{
			Safety: FetchSafetyOptions{ObeyRobots: true},
		})
		if err != nil {
			t.Fatalf("allowed robots path returned error: %v", err)
		}
		if got := response.CSS(".ok").Text(); got != "public" {
			t.Fatalf("allowed response text = %q, want public", got)
		}
		if got := robotsRequests.Load(); got != 2 {
			t.Fatalf("robots requests = %d, want 2 explicit per-request checks", got)
		}
	})

	t.Run("blocks private and configured CIDR targets before dispatch", func(t *testing.T) {
		var requests atomic.Int64
		fetcher := Fetcher{Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			requests.Add(1)
			return nil, errors.New("transport should not be reached")
		})}}

		blocked := []struct {
			name string
			url  string
			opts RequestOptions
		}{
			{
				name: "ipv4 loopback",
				url:  "http://127.0.0.1/private",
				opts: RequestOptions{Safety: FetchSafetyOptions{BlockPrivateNetworks: true}},
			},
			{
				name: "ipv6 loopback",
				url:  "http://[::1]/private",
				opts: RequestOptions{Safety: FetchSafetyOptions{BlockPrivateNetworks: true}},
			},
			{
				name: "configured cidr",
				url:  "http://203.0.113.42/private",
				opts: RequestOptions{Safety: FetchSafetyOptions{BlockedCIDRs: []string{"203.0.113.0/24"}}},
			},
		}

		for _, tt := range blocked {
			t.Run(tt.name, func(t *testing.T) {
				_, err := fetcher.Get(tt.url, tt.opts)
				if !errors.Is(err, ErrPrivateAddressBlocked) {
					t.Fatalf("expected private address blocked error, got %v", err)
				}
				var fetchErr *FetcherError
				if !errors.As(err, &fetchErr) || fetchErr.Kind != FetcherErrorPrivateAddress {
					t.Fatalf("expected private-address FetcherError, got %#v", err)
				}
			})
		}
		if got := requests.Load(); got != 0 {
			t.Fatalf("transport requests = %d, want 0", got)
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestRobotsRuleSpecificity(t *testing.T) {
	robots, err := parseRobotsRules("User-agent: *\nDisallow: /private\nAllow: /private/public\n")
	if err != nil {
		t.Fatalf("parse robots: %v", err)
	}
	if robots.allowed("/private/secret") {
		t.Fatal("expected /private/secret to be blocked")
	}
	if !robots.allowed("/private/public/page") {
		t.Fatal("expected more specific allow rule to win")
	}
	if !robots.allowed("/other") {
		t.Fatal("expected unrelated path to be allowed")
	}
}

func TestParseSafetyCIDRsRejectsInvalidCIDR(t *testing.T) {
	_, err := parseSafetyCIDRs([]string{"not-a-cidr"})
	if err == nil || !strings.Contains(err.Error(), "not-a-cidr") {
		t.Fatalf("expected invalid CIDR error naming input, got %v", err)
	}
}
