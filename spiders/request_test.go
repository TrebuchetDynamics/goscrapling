package spiders_test

import (
	"net/http"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestRequestFingerprint(t *testing.T) {
	t.Run("canonical URL preserves case-sensitive userinfo while normalizing host", func(t *testing.T) {
		upperCredentials := spiders.Request{URL: "https://User:Secret@Example.com/item?b=2&a=1#frag"}
		lowerCredentials := spiders.Request{URL: "https://user:secret@example.com/item?a=1&b=2#other"}

		upperFingerprint, err := upperCredentials.Fingerprint(spiders.FingerprintOptions{})
		if err != nil {
			t.Fatalf("fingerprint upper credentials request: %v", err)
		}
		lowerFingerprint, err := lowerCredentials.Fingerprint(spiders.FingerprintOptions{})
		if err != nil {
			t.Fatalf("fingerprint lower credentials request: %v", err)
		}

		if upperFingerprint == lowerFingerprint {
			t.Fatalf("fingerprints matched; canonicalization lowercased case-sensitive URL userinfo")
		}
	})

	t.Run("include headers preserves non-canonical header map keys", func(t *testing.T) {
		withLowercaseHeader := spiders.Request{
			URL: "https://example.com/item",
			Headers: http.Header{
				"x-trace": []string{"lowercase-value"},
			},
		}
		withoutHeader := spiders.Request{URL: "https://example.com/item"}

		lowercaseFingerprint, err := withLowercaseHeader.Fingerprint(spiders.FingerprintOptions{IncludeHeaders: true})
		if err != nil {
			t.Fatalf("fingerprint lowercase header request: %v", err)
		}
		withoutFingerprint, err := withoutHeader.Fingerprint(spiders.FingerprintOptions{IncludeHeaders: true})
		if err != nil {
			t.Fatalf("fingerprint headerless request: %v", err)
		}

		if lowercaseFingerprint == withoutFingerprint {
			t.Fatalf("fingerprints matched; IncludeHeaders dropped a non-canonical header key")
		}
	})
}
