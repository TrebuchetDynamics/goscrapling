package progress

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AppMap struct {
	Meta    AppMapMeta    `json:"meta"`
	Entries []AppMapEntry `json:"entries"`
}

type AppMapMeta struct {
	Version           string       `json:"version"`
	Upstream          UpstreamMeta `json:"upstream,omitempty"`
	GeneratedMarkdown string       `json:"generated_markdown,omitempty"`
	Py2ManyProbeDir   string       `json:"py2many_probe_dir,omitempty"`
}

type AppMapEntry struct {
	ID                     string      `json:"id"`
	Title                  string      `json:"title"`
	Upstream               []AppMapRef `json:"upstream"`
	FeatureAnchor          string      `json:"feature_anchor,omitempty"`
	BehaviorAtoms          []string    `json:"behavior_atoms,omitempty"`
	GoTarget               string      `json:"go_target,omitempty"`
	ProgressRows           []string    `json:"progress_rows,omitempty"`
	CoverageStatus         string      `json:"coverage_status"`
	TranslationSuitability string      `json:"translation_suitability"`
	StaticReferencePaths   []string    `json:"static_reference_paths,omitempty"`
	Notes                  []string    `json:"notes,omitempty"`
}

type AppMapRef struct {
	Ref     string   `json:"ref"`
	Kind    string   `json:"kind"`
	Symbols []string `json:"symbols,omitempty"`
}

func LoadAppMap(path string) (*AppMap, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var appMap AppMap
	if err := json.Unmarshal(body, &appMap); err != nil {
		return nil, err
	}
	return &appMap, nil
}

func ValidateAppMap(m *AppMap) error {
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
		errs = append(errs, validateAppMapEntry(i, entry, seenIDs)...)
	}
	return errors.Join(errs...)
}

func validateAppMapEntry(index int, entry AppMapEntry, seenIDs map[string]struct{}) []error {
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
		} else if !validAppMapRefKind(ref.Kind) {
			add(fmt.Sprintf("%s invalid upstream kind %q", refPrefix, ref.Kind))
		}
	}
	if !validAppMapCoverageStatus(entry.CoverageStatus) {
		add(fmt.Sprintf("invalid coverage_status %q", entry.CoverageStatus))
	}
	if !validAppMapTranslationSuitability(entry.TranslationSuitability) {
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

func validAppMapRefKind(kind string) bool {
	switch kind {
	case "source", "test", "doc", "asset":
		return true
	default:
		return false
	}
}

func validAppMapCoverageStatus(status string) bool {
	switch status {
	case "covered", "partial", "planned", "vague", "owned", "excluded":
		return true
	default:
		return false
	}
}

func validAppMapTranslationSuitability(status string) bool {
	switch status {
	case "probe_candidate", "manual_rewrite", "not_useful", "not_applicable":
		return true
	default:
		return false
	}
}

func ValidateAppMapCoverage(repoRoot string, m *AppMap) error {
	upstreamRoot := filepath.Join(repoRoot, "references", "Scrapling")
	if _, err := os.Stat(upstreamRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	discovered, err := inventoryAppMapRefs(repoRoot)
	if err != nil {
		return err
	}
	mapped := make(map[string]struct{})
	discoveredSet := make(map[string]struct{}, len(discovered))
	for _, path := range discovered {
		discoveredSet[path] = struct{}{}
	}
	if m != nil {
		for _, entry := range m.Entries {
			for _, ref := range entry.Upstream {
				if trimmed := strings.TrimSpace(ref.Ref); trimmed != "" {
					mapped[filepath.ToSlash(trimmed)] = struct{}{}
				}
			}
		}
	}
	var errs []error
	for _, path := range discovered {
		if _, ok := mapped[path]; !ok {
			errs = append(errs, fmt.Errorf("app map: unmapped upstream ref %s", path))
		}
	}
	mappedRefs := make([]string, 0, len(mapped))
	for path := range mapped {
		mappedRefs = append(mappedRefs, path)
	}
	sort.Strings(mappedRefs)
	for _, path := range mappedRefs {
		if !isInventoriedAppMapRef(path) {
			continue
		}
		if _, ok := discoveredSet[path]; !ok {
			errs = append(errs, fmt.Errorf("app map: stale upstream ref %s", path))
		}
	}
	return errors.Join(errs...)
}

func ValidateAppMapReferences(repoRoot string, m *AppMap, p *Progress) error {
	progressRows := make(map[string]struct{})
	if p != nil {
		for _, phaseKey := range sortedPhaseKeys(p) {
			phase := p.Phases[phaseKey]
			for _, subphaseKey := range sortedSubphaseKeys(phase) {
				subphase := phase.Subphases[subphaseKey]
				for _, item := range subphase.Items {
					if name := strings.TrimSpace(item.Name); name != "" {
						progressRows[name] = struct{}{}
					}
				}
			}
		}
	}

	var errs []error
	if m != nil {
		for _, entry := range m.Entries {
			for _, row := range entry.ProgressRows {
				name := strings.TrimSpace(row)
				if _, ok := progressRows[name]; !ok {
					errs = append(errs, fmt.Errorf("app map: entry %q: unknown progress row %q", entry.ID, row))
				}
			}
			for _, path := range entry.StaticReferencePaths {
				relPath, fullPath, err := cleanStaticReferencePath(repoRoot, path)
				if err != nil {
					errs = append(errs, fmt.Errorf("app map: entry %q: %w", entry.ID, err))
					continue
				}
				info, err := os.Stat(fullPath)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						errs = append(errs, fmt.Errorf("app map: entry %q: missing static reference path %s", entry.ID, relPath))
						continue
					}
					errs = append(errs, fmt.Errorf("app map: entry %q: stat static reference path %s: %w", entry.ID, relPath, err))
					continue
				}
				if !info.Mode().IsRegular() {
					errs = append(errs, fmt.Errorf("app map: entry %q: static reference path is not a regular file %s", entry.ID, relPath))
				}
			}
		}
	}
	return errors.Join(errs...)
}

func cleanStaticReferencePath(repoRoot, path string) (string, string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", "", errors.New("blank static reference path")
	}
	if filepath.IsAbs(trimmed) {
		return "", "", fmt.Errorf("absolute static reference path %s", filepath.ToSlash(trimmed))
	}
	cleaned := filepath.Clean(filepath.FromSlash(trimmed))
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(cleaned), "", fmt.Errorf("static reference path escapes repo root %s", filepath.ToSlash(cleaned))
	}
	relPath := filepath.ToSlash(cleaned)
	return relPath, filepath.Join(repoRoot, cleaned), nil
}

func inventoryAppMapRefs(repoRoot string) ([]string, error) {
	var refs []string
	add := func(path string) {
		refs = append(refs, filepath.ToSlash(path))
	}
	root := filepath.Join(repoRoot, "references", "Scrapling")
	if !exists(root) {
		return refs, nil
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if isInventoriedAppMapRef(rel) {
			add(rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(refs)
	return refs, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isInventoriedAppMapRef(path string) bool {
	path = filepath.ToSlash(path)
	switch {
	case path == "references/Scrapling/scrapling/py.typed":
		return true
	case strings.HasPrefix(path, "references/Scrapling/scrapling/") && strings.HasSuffix(path, ".py"):
		return true
	case strings.HasPrefix(path, "references/Scrapling/tests/") && strings.HasSuffix(path, ".py"):
		return true
	case strings.HasPrefix(path, "references/Scrapling/docs/") && strings.HasSuffix(path, ".md"):
		return true
	case strings.HasPrefix(path, "references/Scrapling/docs/assets/"):
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".png" || ext == ".svg" || ext == ".ico"
	default:
		return false
	}
}

func RenderAppMapMarkdown(m *AppMap) string {
	var b strings.Builder
	b.WriteString("# Upstream Scrapling App Map\n\n")
	if m == nil {
		return b.String()
	}
	renderAppMapMetadata(&b, m.Meta)
	renderAppMapSummary(&b, m.Entries)
	renderAppMapTable(&b, sortedAppMapEntries(m.Entries))
	renderAppMapDetails(&b, sortedAppMapEntries(m.Entries))
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderAppMapMetadata(b *strings.Builder, meta AppMapMeta) {
	b.WriteString("## Source Metadata\n\n")
	fmt.Fprintf(b, "- Version: `%s`\n", meta.Version)
	if meta.Upstream.Name != "" {
		fmt.Fprintf(b, "- Upstream name: `%s`\n", meta.Upstream.Name)
	}
	if meta.Upstream.Repo != "" {
		fmt.Fprintf(b, "- Upstream repo: `%s`\n", meta.Upstream.Repo)
	}
	if meta.Upstream.ObservedCommit != "" {
		fmt.Fprintf(b, "- Observed commit: `%s`\n", meta.Upstream.ObservedCommit)
	}
	if meta.Upstream.ObservedRelease != "" {
		fmt.Fprintf(b, "- Observed release: `%s`\n", meta.Upstream.ObservedRelease)
	}
	if meta.Upstream.LocalCheckout != "" {
		fmt.Fprintf(b, "- Local checkout: `%s`\n", meta.Upstream.LocalCheckout)
	}
	if meta.GeneratedMarkdown != "" {
		fmt.Fprintf(b, "- Generated markdown: `%s`\n", meta.GeneratedMarkdown)
	}
	if meta.Py2ManyProbeDir != "" {
		fmt.Fprintf(b, "- py2many probe dir: `%s`\n", meta.Py2ManyProbeDir)
	}
	b.WriteString("\n")
}

func renderAppMapSummary(b *strings.Builder, entries []AppMapEntry) {
	counts := make(map[string]int)
	for _, entry := range entries {
		counts[entry.CoverageStatus]++
	}
	b.WriteString("## Coverage Summary\n\n")
	b.WriteString("| Status | Count |\n")
	b.WriteString("|---|---:|\n")
	for _, status := range []string{"covered", "partial", "planned", "vague", "owned", "excluded"} {
		if counts[status] > 0 {
			fmt.Fprintf(b, "| `%s` | %d |\n", status, counts[status])
		}
	}
	b.WriteString("\n")
}

func renderAppMapTable(b *strings.Builder, entries []AppMapEntry) {
	b.WriteString("## Entries\n\n")
	b.WriteString("| Entry | Status | Feature anchor | Go target | Progress rows | Translation suitability | Upstream refs |\n")
	b.WriteString("|---|---|---|---|---|---|---:|\n")
	for _, entry := range entries {
		fmt.Fprintf(b, "| %s | `%s` | `%s` | `%s` | %s | `%s` | %d |\n",
			escapeCell(entry.Title),
			escapeCell(entry.CoverageStatus),
			escapeCell(entry.FeatureAnchor),
			escapeCell(entry.GoTarget),
			escapeCell(joinCodeOrDash(entry.ProgressRows)),
			escapeCell(entry.TranslationSuitability),
			len(entry.Upstream),
		)
	}
	b.WriteString("\n")
}

func renderAppMapDetails(b *strings.Builder, entries []AppMapEntry) {
	for _, entry := range entries {
		fmt.Fprintf(b, "## %s\n\n", entry.Title)
		fmt.Fprintf(b, "- ID: `%s`\n", entry.ID)
		if len(entry.Upstream) > 0 {
			b.WriteString("- Upstream refs:\n")
			for _, ref := range entry.Upstream {
				fmt.Fprintf(b, "  - `%s` (`%s`)", ref.Ref, ref.Kind)
				if len(ref.Symbols) > 0 {
					fmt.Fprintf(b, " symbols: %s", joinCodeOrDash(ref.Symbols))
				}
				b.WriteString("\n")
			}
		}
		if len(entry.StaticReferencePaths) > 0 {
			fmt.Fprintf(b, "- Static reference paths: %s\n", joinCodeOrDash(entry.StaticReferencePaths))
		}
		if len(entry.BehaviorAtoms) > 0 {
			fmt.Fprintf(b, "- Behavior atoms: %s\n", joinOrDash(entry.BehaviorAtoms))
		}
		if len(entry.Notes) > 0 {
			fmt.Fprintf(b, "- Notes: %s\n", joinOrDash(entry.Notes))
		}
		b.WriteString("\n")
	}
}

func sortedAppMapEntries(entries []AppMapEntry) []AppMapEntry {
	out := append([]AppMapEntry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ID == out[j].ID {
			return out[i].Title < out[j].Title
		}
		return out[i].ID < out[j].ID
	})
	return out
}
