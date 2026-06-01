package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/schema"
)

func Validate(m *schema.AppMap) error {
	if m == nil {
		return errors.New("app map: nil app map")
	}
	var errs []error
	if strings.TrimSpace(m.Meta.Version) == "" {
		errs = append(errs, errors.New("app map: meta.version is required"))
	}
	if len(m.Entries) == 0 {
		errs = append(errs, errors.New("app map: entries are required"))
	}
	seenIDs := make(map[string]struct{})
	for i, entry := range m.Entries {
		errs = append(errs, validateEntry(i, entry, seenIDs)...)
	}
	return errors.Join(errs...)
}

func validateEntry(index int, entry schema.AppMapEntry, seenIDs map[string]struct{}) []error {
	var errs []error
	prefix := func() string {
		return fmt.Sprintf("app map: entry[%d] (%q): ", index, entry.ID)
	}
	add := func(message string) {
		errs = append(errs, errors.New(prefix()+message))
	}
	id := strings.TrimSpace(entry.ID)
	if id == "" {
		add("missing id")
	} else if _, exists := seenIDs[id]; exists {
		add("duplicate entry id")
	} else {
		seenIDs[id] = struct{}{}
	}
	if strings.TrimSpace(entry.Title) == "" {
		add("missing title")
	}
	if len(entry.Upstream) == 0 {
		add("missing upstream refs")
	}
	for i, ref := range entry.Upstream {
		refPrefix := fmt.Sprintf("upstream[%d]", i)
		if strings.TrimSpace(ref.Ref) == "" {
			add(refPrefix + " missing ref")
		}
		if strings.TrimSpace(ref.Kind) == "" {
			add(refPrefix + " missing kind")
		} else if !validRefKind(ref.Kind) {
			add(fmt.Sprintf("%s invalid upstream kind %q", refPrefix, ref.Kind))
		}
	}
	if !validCoverageStatus(entry.CoverageStatus) {
		add(fmt.Sprintf("invalid coverage_status %q", entry.CoverageStatus))
	}
	if !validTranslationSuitability(entry.TranslationSuitability) {
		add(fmt.Sprintf("invalid translation_suitability %q", entry.TranslationSuitability))
	}
	if entry.CoverageStatus != "excluded" {
		if strings.TrimSpace(entry.FeatureAnchor) == "" {
			add("missing feature_anchor")
		}
		if strings.TrimSpace(entry.GoTarget) == "" {
			add("missing go_target")
		}
		if len(entry.ProgressRows) == 0 {
			add("missing progress_rows")
		}
	}
	return errs
}

func validRefKind(kind string) bool {
	switch kind {
	case "source", "test", "doc", "asset":
		return true
	default:
		return false
	}
}

func validCoverageStatus(status string) bool {
	switch status {
	case "covered", "partial", "planned", "vague", "owned", "excluded":
		return true
	default:
		return false
	}
}

func validTranslationSuitability(status string) bool {
	switch status {
	case "probe_candidate", "manual_rewrite", "not_useful", "not_applicable":
		return true
	default:
		return false
	}
}
