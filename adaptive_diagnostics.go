package goscrapling

import (
	"context"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	DiagnosticBelowThreshold = "below-threshold"
	DiagnosticNoCandidates   = "no-candidates"
	DiagnosticMissingStore   = "missing-store-record"
)

type DiagnosticOptions struct {
	Percentage float64
	Domain     string
}

type AdaptiveDiagnostic struct {
	Key               Key
	Target            Fingerprint
	MinScore          float64
	CandidateCount    int
	Candidates        []CandidateDiagnostic
	Best              CandidateDiagnostic
	Accepted          bool
	FailureReason     string
	FingerprintFields []FingerprintFieldDiagnostic
}

type CandidateDiagnostic struct {
	Element     *Element
	Fingerprint Fingerprint
	Score       float64
	Components  []ScoreComponent
}

type FingerprintFieldDiagnostic struct {
	Name         string
	UpstreamName string
	Present      bool
}

func (d *Document) DiagnoseRelocate(ctx context.Context, identifier string, opts DiagnosticOptions) (AdaptiveDiagnostic, error) {
	if d == nil || d.store == nil {
		return AdaptiveDiagnostic{}, ErrMissingStore
	}
	if ctx == nil {
		ctx = context.Background()
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return AdaptiveDiagnostic{}, ErrEmptyIdentifier
	}

	minScore, err := minScoreFromPercentage(opts.Percentage)
	if err != nil {
		return AdaptiveDiagnostic{}, err
	}

	key := Key{
		Domain:     selectorDomain(d.domain, opts.Domain),
		Identifier: identifier,
	}
	target, ok, err := d.store.Load(ctx, key)
	if err != nil {
		return AdaptiveDiagnostic{}, err
	}

	diagnostic := AdaptiveDiagnostic{
		Key:               key,
		Target:            target,
		MinScore:          minScore,
		FingerprintFields: fingerprintFieldDiagnostics(target),
	}
	if !ok {
		diagnostic.FailureReason = DiagnosticMissingStore
		return diagnostic, nil
	}
	if d.query == nil {
		diagnostic.FailureReason = DiagnosticNoCandidates
		return diagnostic, nil
	}

	d.query.Find("*").Each(func(_ int, selection *goquery.Selection) {
		for _, node := range selection.Nodes {
			fingerprint := fingerprintNode(node)
			components := scoreFingerprintComponents(fingerprint, target)
			score := sumScoreComponents(components)
			diagnostic.Candidates = append(diagnostic.Candidates, CandidateDiagnostic{
				Element:     &Element{doc: d, node: node},
				Fingerprint: fingerprint,
				Score:       score,
				Components:  components,
			})
		}
	})
	diagnostic.CandidateCount = len(diagnostic.Candidates)
	if diagnostic.CandidateCount == 0 {
		diagnostic.FailureReason = DiagnosticNoCandidates
		return diagnostic, nil
	}

	sort.SliceStable(diagnostic.Candidates, func(i, j int) bool {
		return diagnostic.Candidates[i].Score > diagnostic.Candidates[j].Score
	})
	diagnostic.Best = diagnostic.Candidates[0]
	if diagnostic.Best.Score >= minScore {
		diagnostic.Accepted = true
		return diagnostic, nil
	}

	diagnostic.FailureReason = DiagnosticBelowThreshold
	return diagnostic, nil
}

func sumScoreComponents(components []ScoreComponent) float64 {
	var score float64
	for _, component := range components {
		score += component.Contribution
	}
	return score
}

func fingerprintFieldDiagnostics(fp Fingerprint) []FingerprintFieldDiagnostic {
	return []FingerprintFieldDiagnostic{
		{Name: "Tag", UpstreamName: "tag", Present: fp.Tag != ""},
		{Name: "Attributes", UpstreamName: "attributes", Present: fp.Attributes != nil},
		{Name: "Text", UpstreamName: "text", Present: true},
		{Name: "PathTags", UpstreamName: "path", Present: fp.PathTags != nil},
		{Name: "ParentTag", UpstreamName: "parent_name", Present: fp.ParentTag != ""},
		{Name: "ParentAttributes", UpstreamName: "parent_attribs", Present: fp.ParentAttributes != nil},
		{Name: "ParentText", UpstreamName: "parent_text", Present: true},
		{Name: "SiblingTags", UpstreamName: "siblings", Present: fp.SiblingTags != nil},
		{Name: "ChildrenTags", UpstreamName: "children", Present: fp.ChildrenTags != nil},
	}
}
