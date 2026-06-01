package markdown

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/appmap/schema"
	"github.com/TrebuchetDynamics/goscrapling/internal/progress/formatting"
)

func Render(m *schema.AppMap) string {
	var b strings.Builder
	b.WriteString("# Upstream Scrapling App Map\n\n")
	if m == nil {
		return b.String()
	}
	renderMetadata(&b, m.Meta)
	renderSummary(&b, m.Entries)
	renderTable(&b, sortedEntries(m.Entries))
	renderDetails(&b, sortedEntries(m.Entries))
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderMetadata(b *strings.Builder, meta schema.AppMapMeta) {
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

func renderSummary(b *strings.Builder, entries []schema.AppMapEntry) {
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

func renderTable(b *strings.Builder, entries []schema.AppMapEntry) {
	b.WriteString("## Entries\n\n")
	b.WriteString("| Entry | Status | Feature anchor | Go target | Progress rows | Translation suitability | Upstream refs |\n")
	b.WriteString("|---|---|---|---|---|---|---:|\n")
	for _, entry := range entries {
		fmt.Fprintf(b, "| %s | `%s` | `%s` | `%s` | %s | `%s` | %d |\n",
			formatting.EscapeCell(entry.Title),
			formatting.EscapeCell(entry.CoverageStatus),
			formatting.EscapeCell(entry.FeatureAnchor),
			formatting.EscapeCell(entry.GoTarget),
			formatting.EscapeCell(formatting.JoinCodeOrDash(entry.ProgressRows)),
			formatting.EscapeCell(entry.TranslationSuitability),
			len(entry.Upstream),
		)
	}
	b.WriteString("\n")
}

func renderDetails(b *strings.Builder, entries []schema.AppMapEntry) {
	for _, entry := range entries {
		fmt.Fprintf(b, "## %s\n\n", entry.Title)
		fmt.Fprintf(b, "- ID: `%s`\n", entry.ID)
		if len(entry.Upstream) > 0 {
			b.WriteString("- Upstream refs:\n")
			for _, ref := range entry.Upstream {
				fmt.Fprintf(b, "  - `%s` (`%s`)", ref.Ref, ref.Kind)
				if len(ref.Symbols) > 0 {
					fmt.Fprintf(b, " symbols: %s", formatting.JoinCodeOrDash(ref.Symbols))
				}
				b.WriteString("\n")
			}
		}
		if len(entry.StaticReferencePaths) > 0 {
			fmt.Fprintf(b, "- Static reference paths: %s\n", formatting.JoinCodeOrDash(entry.StaticReferencePaths))
		}
		if len(entry.BehaviorAtoms) > 0 {
			fmt.Fprintf(b, "- Behavior atoms: %s\n", formatting.JoinOrDash(entry.BehaviorAtoms))
		}
		if len(entry.Notes) > 0 {
			fmt.Fprintf(b, "- Notes: %s\n", formatting.JoinOrDash(entry.Notes))
		}
		b.WriteString("\n")
	}
}

func sortedEntries(entries []schema.AppMapEntry) []schema.AppMapEntry {
	out := append([]schema.AppMapEntry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ID == out[j].ID {
			return out[i].Title < out[j].Title
		}
		return out[i].ID < out[j].ID
	})
	return out
}
