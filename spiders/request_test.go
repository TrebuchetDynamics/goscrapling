package spiders_test

import (
	"net/http"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestRequestFingerprint(t *testing.T) {
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
