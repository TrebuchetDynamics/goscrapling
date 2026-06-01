package spiders

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestCollectLinkAttributeCandidatesExposesDOMOrder(t *testing.T) {
	root, err := html.Parse(strings.NewReader(`<main><a href="/one" data-url="/two"><span><area href="/three"></area></span></a><link href="/skip"></main>`))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	candidates := collectLinkAttributeCandidates(root, stringSet([]string{"a", "area"}), stringSet([]string{"href", "data-url"}))
	want := []linkAttributeCandidate{
		{tag: "a", attr: "href", raw: "/one"},
		{tag: "a", attr: "data-url", raw: "/two"},
		{tag: "area", attr: "href", raw: "/three"},
	}
	if !reflect.DeepEqual(candidates, want) {
		t.Fatalf("candidates = %#v, want %#v", candidates, want)
	}
}

func TestNilLinkExtractorMatchHelpersAreSafe(t *testing.T) {
	var extractor *LinkExtractor
	if extractor.Matches("https://example.com/page") {
		t.Fatal("nil extractor Matches should reject candidates")
	}
	if got, ok := extractor.MatchURL("https://example.com/page"); ok || got != "" {
		t.Fatalf("nil extractor MatchURL = %q, %v; want empty false", got, ok)
	}
	if got, ok := extractor.MatchURLFrom("https://example.com/", "/page"); ok || got != "" {
		t.Fatalf("nil extractor MatchURLFrom = %q, %v; want empty false", got, ok)
	}
}

func TestLinkExtractorPrepareCandidateExposesURLStages(t *testing.T) {
	extractor, err := NewLinkExtractor(LinkExtractorOptions{
		Process: func(raw string) (string, bool) {
			if raw != "https://example.com/old?b=2&a=1#drop" {
				t.Fatalf("process saw %q, want resolved candidate URL", raw)
			}
			return "../new?z=9&a=1#drop", true
		},
	})
	if err != nil {
		t.Fatalf("NewLinkExtractor returned error: %v", err)
	}

	candidate, ok, err := extractor.prepareCandidate("https://example.com/base/index.html", " /old?b=2&a=1#drop ")
	if err != nil {
		t.Fatalf("prepareCandidate returned error: %v", err)
	}
	if !ok {
		t.Fatal("prepareCandidate dropped candidate")
	}
	if candidate.candidate.raw != "/old?b=2&a=1#drop" {
		t.Fatalf("raw candidate = %q, want stripped raw URL", candidate.candidate.raw)
	}
	if candidate.candidate.resolved != "https://example.com/old?b=2&a=1#drop" {
		t.Fatalf("resolved candidate = %q", candidate.candidate.resolved)
	}
	if candidate.url != "https://example.com/new?a=1&z=9" {
		t.Fatalf("prepared URL = %q, want processed and canonicalized URL", candidate.url)
	}
}

func TestLinkCandidatePipelineFiltersCanonicalizedURL(t *testing.T) {
	var filtered string
	config := linkCandidateConfig{
		strip:        true,
		process:      defaultLinkProcess,
		canonicalize: true,
		passes: func(rawURL string) bool {
			filtered = rawURL
			return true
		},
	}

	result := config.prepare("https://Example.COM/base/index.html", " /keep/../file?b=2&a=1#drop ")
	if result.err != nil {
		t.Fatalf("prepare returned error: %v", result.err)
	}
	if !result.ok {
		t.Fatalf("prepare dropped candidate with reason %q", result.reason)
	}
	want := "https://example.com/file?a=1&b=2"
	if result.candidate.url != want {
		t.Fatalf("prepared URL = %q, want %q", result.candidate.url, want)
	}
	if filtered != want {
		t.Fatalf("filter saw %q, want canonical URL %q", filtered, want)
	}
}

func TestLinkCandidatePipelineCanExposePreCanonicalFilterInput(t *testing.T) {
	var filtered string
	config := linkCandidateConfig{
		strip:        true,
		process:      defaultLinkProcess,
		canonicalize: false,
		passes: func(rawURL string) bool {
			filtered = rawURL
			return true
		},
	}

	result := config.prepare("https://Example.COM/base/index.html", " /keep/../file?b=2&a=1#keep ")
	if result.err != nil {
		t.Fatalf("prepare returned error: %v", result.err)
	}
	if !result.ok {
		t.Fatalf("prepare dropped candidate with reason %q", result.reason)
	}
	want := "https://Example.COM/file?b=2&a=1#keep"
	if result.candidate.url != want {
		t.Fatalf("prepared URL = %q, want %q", result.candidate.url, want)
	}
	if filtered != want {
		t.Fatalf("filter saw %q, want uncanonicalized URL %q", filtered, want)
	}
}

func TestLinkExtractorRejectsDeniedExtensionsWithPathParameters(t *testing.T) {
	extractor, err := NewLinkExtractor(LinkExtractorOptions{})
	if err != nil {
		t.Fatalf("NewLinkExtractor returned error: %v", err)
	}

	result := extractor.prepareCandidateDiagnostic("https://example.com/index.html", "/reports/annual.pdf;download?cache=1")
	if result.ok {
		t.Fatalf("candidate unexpectedly survived: %#v", result.candidate)
	}
	if result.reason != linkDropFiltered {
		t.Fatalf("drop reason = %q, want %q", result.reason, linkDropFiltered)
	}
}

func TestLinkExtractorPrepareCandidateDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		opts   LinkExtractorOptions
		reason linkDropReason
	}{
		{
			name:   "empty raw",
			raw:    "   ",
			reason: linkDropEmptyRaw,
		},
		{
			name:   "invalid raw URL",
			raw:    "http://[::1",
			reason: linkDropInvalidRaw,
		},
		{
			name: "process rejected",
			raw:  "/drop",
			opts: LinkExtractorOptions{Process: func(string) (string, bool) {
				return "", false
			}},
			reason: linkDropProcessRejected,
		},
		{
			name: "invalid processed URL",
			raw:  "/bad",
			opts: LinkExtractorOptions{Process: func(string) (string, bool) {
				return "\n/bad", true
			}},
			reason: linkDropInvalidProcessed,
		},
		{
			name:   "filtered extension",
			raw:    "/report.pdf",
			reason: linkDropFiltered,
		},
	}

	gotReasons := make([]linkDropReason, 0, len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := NewLinkExtractor(tt.opts)
			if err != nil {
				t.Fatalf("NewLinkExtractor returned error: %v", err)
			}
			result := extractor.prepareCandidateDiagnostic("https://example.com/base/index.html", tt.raw)
			if result.ok {
				t.Fatalf("candidate unexpectedly survived: %#v", result.candidate)
			}
			if result.err != nil {
				t.Fatalf("dropped candidate returned fatal error: %v", result.err)
			}
			if result.reason != tt.reason {
				t.Fatalf("drop reason = %q, want %q", result.reason, tt.reason)
			}
			gotReasons = append(gotReasons, result.reason)
		})
	}

	wantReasons := []linkDropReason{linkDropEmptyRaw, linkDropInvalidRaw, linkDropProcessRejected, linkDropInvalidProcessed, linkDropFiltered}
	if !reflect.DeepEqual(gotReasons, wantReasons) {
		t.Fatalf("drop reasons = %#v, want %#v", gotReasons, wantReasons)
	}
}
