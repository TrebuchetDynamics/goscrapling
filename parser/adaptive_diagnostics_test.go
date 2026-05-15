package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/core/storage"
)

func TestAdaptiveDiagnostics(t *testing.T) {
	t.Run("explains top candidate scores", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDiagnosticsDocument(t, `<section class="products"><article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article></section>`, store)
		element, _ := before.CSS("#p1").First()
		if err := before.Save(ctx, element, "featured"); err != nil {
			t.Fatalf("Save: %v", err)
		}

		after := mustParseAdaptiveDiagnosticsDocument(t, `<section class="products"><article class="product new" data-id="p1"><div><h2>Product 1</h2><p>Description 1</p></div></article><article class="product" data-id="p2">Other</article></section>`, store)
		diagnostic, err := after.DiagnoseRelocate(ctx, "featured", DiagnosticOptions{})
		if err != nil {
			t.Fatalf("DiagnoseRelocate: %v", err)
		}

		if !diagnostic.Accepted {
			t.Fatalf("expected accepted diagnostic, got reason %q score %.3f", diagnostic.FailureReason, diagnostic.Best.Score)
		}
		if diagnostic.CandidateCount == 0 || len(diagnostic.Candidates) == 0 {
			t.Fatalf("expected scored candidates, got count=%d len=%d", diagnostic.CandidateCount, len(diagnostic.Candidates))
		}
		if got := diagnostic.Best.Element.Text(); got != "Product 1 Description 1" {
			t.Fatalf("expected best relocated product, got %q", got)
		}
		if diagnostic.Best.Score < diagnostic.MinScore {
			t.Fatalf("expected best score above threshold, score=%.3f threshold=%.3f", diagnostic.Best.Score, diagnostic.MinScore)
		}
		assertScoreComponent(t, diagnostic.Best.Components, "tag")
		assertScoreComponent(t, diagnostic.Best.Components, "text")
		assertScoreComponent(t, diagnostic.Best.Components, "attributes.keys")
		assertScoreComponent(t, diagnostic.Best.Components, "attributes.values")
		assertScoreComponent(t, diagnostic.Best.Components, "parent")
		assertScoreComponent(t, diagnostic.Best.Components, "siblings")
		assertScoreComponent(t, diagnostic.Best.Components, "path")
		assertScoreComponent(t, diagnostic.Best.Components, "children")
	})

	t.Run("reports threshold failures", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDiagnosticsDocument(t, `<article class="product" id="p1">Product 1</article>`, store)
		element, _ := before.CSS("#p1").First()
		if err := before.Save(ctx, element, "featured"); err != nil {
			t.Fatalf("Save: %v", err)
		}

		after := mustParseAdaptiveDiagnosticsDocument(t, `<article class="product" data-id="p1"><span>Product 1</span></article>`, store)
		diagnostic, err := after.DiagnoseRelocate(ctx, "featured", DiagnosticOptions{Percentage: 99})
		if err != nil {
			t.Fatalf("DiagnoseRelocate: %v", err)
		}

		if diagnostic.Accepted {
			t.Fatal("expected high threshold diagnostic to be rejected")
		}
		if diagnostic.FailureReason != DiagnosticBelowThreshold {
			t.Fatalf("expected below-threshold reason, got %q", diagnostic.FailureReason)
		}
		if diagnostic.Best.Score >= diagnostic.MinScore {
			t.Fatalf("expected best score below threshold, score=%.3f threshold=%.3f", diagnostic.Best.Score, diagnostic.MinScore)
		}
	})

	t.Run("documents upstream fingerprint fields", func(t *testing.T) {
		ctx := context.Background()
		store := storage.NewMemoryStore()
		before := mustParseAdaptiveDiagnosticsDocument(t, `<main><section class="products"><article class="product" id="p1"><h2>Product 1</h2><p>Description 1</p></article><aside>Related</aside></section></main>`, store)
		element, _ := before.CSS("#p1").First()
		if err := before.Save(ctx, element, "featured"); err != nil {
			t.Fatalf("Save: %v", err)
		}

		after := mustParseAdaptiveDiagnosticsDocument(t, `<main><section class="products"><article class="product" data-id="p1"><h2>Product 1</h2><p>Description 1</p></article><aside>Related</aside></section></main>`, store)
		diagnostic, err := after.DiagnoseRelocate(ctx, "featured", DiagnosticOptions{})
		if err != nil {
			t.Fatalf("DiagnoseRelocate: %v", err)
		}

		fields := map[string]FingerprintFieldDiagnostic{}
		for _, field := range diagnostic.FingerprintFields {
			fields[field.UpstreamName] = field
		}
		for _, name := range []string{"tag", "attributes", "text", "path", "parent_name", "parent_attribs", "parent_text", "siblings", "children"} {
			field, ok := fields[name]
			if !ok {
				t.Fatalf("expected upstream fingerprint field %q to be documented", name)
			}
			if !field.Present {
				t.Fatalf("expected upstream fingerprint field %q to be present in Go fingerprint", name)
			}
		}
	})
}

func mustParseAdaptiveDiagnosticsDocument(t *testing.T, body string, store Store) *Document {
	t.Helper()
	doc, err := Parse(strings.NewReader(body), ParseOptions{URL: "https://example.com/products", Store: store})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

func assertScoreComponent(t *testing.T, components []ScoreComponent, name string) {
	t.Helper()
	for _, component := range components {
		if component.Name == name {
			if component.Weight <= 0 {
				t.Fatalf("expected component %q to have positive weight", name)
			}
			if component.Score < 0 || component.Score > 1 {
				t.Fatalf("expected component %q score in [0,1], got %.3f", name, component.Score)
			}
			return
		}
	}
	t.Fatalf("expected score component %q in %#v", name, components)
}
