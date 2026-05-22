package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress"
)

const parityScorecardPath = "docs/research/parity-scorecard.md"

type scorecardArea struct {
	Name     string
	Owners   []string
	Fixtures []string
}

type scorecardStats struct {
	Complete   int
	InProgress int
	Planned    int
}

func writeScorecard(stdout io.Writer, root string) error {
	p, err := loadValid(root)
	if err != nil {
		return err
	}
	benchmarks, err := discoverBenchmarkFixtures(filepath.Join(root, "benchmarks"))
	if err != nil {
		return err
	}
	body := renderParityScorecard(p, benchmarks)
	target := filepath.Join(root, parityScorecardPath)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create scorecard directory: %w", err)
	}
	if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", parityScorecardPath, err)
	}
	_, err = fmt.Fprintf(stdout, "progress: wrote %s\n", parityScorecardPath)
	return err
}

func renderParityScorecard(p *progress.Progress, benchmarks []string) string {
	areas := []scorecardArea{
		{Name: "Parser and selectors", Owners: []string{"parser"}, Fixtures: []string{"BenchmarkParserNestedText"}},
		{Name: "Static fetcher and response", Owners: []string{"fetcher"}, Fixtures: []string{"BenchmarkStaticFetcherLocalResponse"}},
		{Name: "Spider runtime", Owners: []string{"spider"}, Fixtures: []string{"BenchmarkSpiderSchedulerFingerprint"}},
		{Name: "CLI shell and extract commands", Owners: []string{"cli"}, Fixtures: []string{"BenchmarkCLIExtractFixture"}},
	}
	benchmarkSet := make(map[string]bool, len(benchmarks))
	for _, name := range benchmarks {
		benchmarkSet[name] = true
	}

	var b strings.Builder
	b.WriteString("# goscrapling Parity Scorecard\n\n")
	b.WriteString("Generated from `progress.json` and benchmark fixture names under `benchmarks/`.\n")
	b.WriteString("This scorecard is evidence for current Go port coverage; it is not a complete Scrapling parity claim.\n\n")

	b.WriteString("## Upstream Benchmark Anchors\n\n")
	b.WriteString("Source: `references/Scrapling/docs/benchmarks.md`.\n\n")
	b.WriteString("- Text Extraction Speed Test: upstream reports Scrapling at 2.02 ms for 5000 nested elements.\n")
	b.WriteString("- Element Similarity & Text Search Performance: upstream reports Scrapling at 2.39 ms.\n")
	b.WriteString("- Local Go timings should be captured separately with `go test ./benchmarks -bench . -benchmem`.\n\n")

	b.WriteString("## Coverage Scorecard\n\n")
	b.WriteString("| Area | Status | Complete | In progress | Planned | Benchmark fixture |\n")
	b.WriteString("|---|---:|---:|---:|---:|---|\n")
	for _, area := range areas {
		stats := statsForOwners(p, area.Owners)
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d | %s |\n",
			area.Name,
			areaStatus(stats),
			stats.Complete,
			stats.InProgress,
			stats.Planned,
			strings.Join(availableFixtures(area.Fixtures, benchmarkSet), ", "),
		)
	}

	b.WriteString("\n## Benchmark Fixtures\n\n")
	if len(benchmarks) == 0 {
		b.WriteString("No benchmark fixtures found under `benchmarks/`.\n")
	} else {
		for _, name := range benchmarks {
			fmt.Fprintf(&b, "- `%s`\n", name)
		}
	}
	b.WriteString("\nNo live network, live browser, or live LLM is required: fixtures use in-memory HTML, `httptest`, deterministic spider scheduler data, and local CLI output files.\n")
	return b.String()
}

func statsForOwners(p *progress.Progress, owners []string) scorecardStats {
	ownerSet := make(map[string]bool, len(owners))
	for _, owner := range owners {
		ownerSet[owner] = true
	}
	var stats scorecardStats
	phaseKeys := make([]string, 0, len(p.Phases))
	for key := range p.Phases {
		phaseKeys = append(phaseKeys, key)
	}
	sort.Strings(phaseKeys)
	for _, phaseKey := range phaseKeys {
		phase := p.Phases[phaseKey]
		subphaseKeys := make([]string, 0, len(phase.Subphases))
		for key := range phase.Subphases {
			subphaseKeys = append(subphaseKeys, key)
		}
		sort.Strings(subphaseKeys)
		for _, subphaseKey := range subphaseKeys {
			for _, item := range phase.Subphases[subphaseKey].Items {
				if item.Contract == "" || !ownerSet[item.ExecutionOwner] {
					continue
				}
				switch item.Status {
				case progress.StatusComplete:
					stats.Complete++
				case progress.StatusInProgress:
					stats.InProgress++
				case progress.StatusPlanned:
					stats.Planned++
				}
			}
		}
	}
	return stats
}

func areaStatus(stats scorecardStats) string {
	switch {
	case stats.Complete == 0 && stats.InProgress == 0:
		return "planned"
	case stats.Planned > 0 || stats.InProgress > 0:
		return "partial"
	default:
		return "covered"
	}
}

func availableFixtures(wanted []string, benchmarkSet map[string]bool) []string {
	out := make([]string, 0, len(wanted))
	for _, name := range wanted {
		if benchmarkSet[name] {
			out = append(out, "`"+name+"`")
		}
	}
	if len(out) == 0 {
		return []string{"-"}
	}
	return out
}

func discoverBenchmarkFixtures(root string) ([]string, error) {
	benchmarkPattern := regexp.MustCompile(`func\s+(Benchmark[A-Za-z0-9_]+)\s*\(`)
	seen := map[string]bool{}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat benchmarks: %w", err)
	}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range benchmarkPattern.FindAllStringSubmatch(string(body), -1) {
			seen[match[1]] = true
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("discover benchmark fixtures: %w", err)
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
