package spiders

import (
	"net/http"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
)

func TestBrowserRequestOptionsSnapshotCookies(t *testing.T) {
	cookie := &http.Cookie{Name: "session", Value: "original"}
	request := WithBrowserRequestOptions(Request{URL: "https://example.com"}, browser.BrowserOptions{
		Cookies: []*http.Cookie{cookie},
	})

	cookie.Value = "mutated"
	stored := browserRequestOptionsFromMeta(request.Meta)
	if len(stored.Cookies) != 1 {
		t.Fatalf("stored cookies = %#v, want one cookie", stored.Cookies)
	}
	if stored.Cookies[0].Value != "original" {
		t.Fatalf("stored cookie value = %q, want snapshot value %q", stored.Cookies[0].Value, "original")
	}

	stored.Cookies[0].Value = "changed-after-read"
	storedAgain := browserRequestOptionsFromMeta(request.Meta)
	if storedAgain.Cookies[0].Value != "original" {
		t.Fatalf("second read cookie value = %q, want metadata isolated from returned options", storedAgain.Cookies[0].Value)
	}
}

func TestMergeSpiderBrowserOptionsDoesNotAliasOverrideCookies(t *testing.T) {
	overrideCookie := &http.Cookie{Name: "request", Value: "override"}
	merged := mergeSpiderBrowserOptions(browser.BrowserOptions{}, browser.BrowserOptions{
		Cookies: []*http.Cookie{overrideCookie},
	})

	overrideCookie.Value = "mutated"
	if len(merged.Cookies) != 1 {
		t.Fatalf("merged cookies = %#v, want one cookie", merged.Cookies)
	}
	if merged.Cookies[0].Value != "override" {
		t.Fatalf("merged cookie value = %q, want snapshot value %q", merged.Cookies[0].Value, "override")
	}
}
