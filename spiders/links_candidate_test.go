package spiders

import "testing"

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
