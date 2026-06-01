package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress"
)

const usage = "usage: progress [--repo-root <path>] {validate|write|map-validate|map-write|scorecard}"

var errParse = errors.New("parse error")

func main() {
	if err := run(os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, errParse) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(stdout, stderr io.Writer, args []string) error {
	args, root, err := resolveRepoRoot(args)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("%w\n%s", errParse, usage)
	}
	switch args[0] {
	case "--help", "-h", "help":
		_, err := fmt.Fprintln(stdout, usage)
		return err
	case "validate":
		p, err := loadValid(root)
		if err != nil {
			return err
		}
		stats := p.Stats()
		_, err = fmt.Fprintf(stdout, "progress: validated %d phases, %d items\n", stats.Phases.Total, stats.Items.Total)
		return err
	case "write":
		return writeDocs(stdout, root)
	case "map-validate":
		appMap, err := loadValidAppMap(root)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "app-map: validated %d entries\n", len(appMap.Entries))
		return err
	case "map-write":
		return writeAppMap(stdout, root)
	case "scorecard":
		return writeScorecard(stdout, root)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown progress command %q\n", args[0])
		return fmt.Errorf("%w\n%s", errParse, usage)
	}
}

func resolveRepoRoot(args []string) ([]string, string, error) {
	out := make([]string, 0, len(args))
	root := os.Getenv("REPO_ROOT")
	for i := 0; i < len(args); i++ {
		if args[i] == "--repo-root" {
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("%w: --repo-root requires a value\n%s", errParse, usage)
			}
			root = args[i+1]
			i++
			continue
		}
		out = append(out, args[i])
	}
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}
		root = cwd
	}
	return out, root, nil
}

func writeDocs(stdout io.Writer, root string) error {
	p, err := loadValid(root)
	if err != nil {
		return err
	}
	for _, target := range []struct {
		path string
		kind string
		body string
	}{
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/handoff/builder-loop-handoff.md", kind: "builder-loop-handoff", body: progress.RenderBuilderLoopHandoff(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/assignable/agent-queue.md", kind: "agent-queue", body: progress.RenderAgentQueue(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/assignable/next-slices.md", kind: "next-slices", body: progress.RenderNextSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked/blocked-slices.md", kind: "blocked-slices", body: progress.RenderBlockedSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/agent-queue.md", kind: "agent-queue", body: progress.RenderAgentQueue(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/next-slices.md", kind: "next-slices", body: progress.RenderNextSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/queue/blocked-slices.md", kind: "blocked-slices", body: progress.RenderBlockedSlices(p)},
		{path: "docs/content/building-goscrapling/builder-loop/surfaces/cleanup/umbrella-cleanup.md", kind: "umbrella-cleanup", body: progress.RenderUmbrellaCleanup(p)},
	} {
		if err := rewriteMarker(filepath.Join(root, target.path), target.kind, target.body); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "progress: regenerated %s\n", target.path)
	}
	return nil
}

func loadValid(root string) (*progress.Progress, error) {
	p, err := progress.Load(filepath.Join(root, "docs/content/building-goscrapling/architecture_plan/progress.json"))
	if err != nil {
		return nil, fmt.Errorf("load progress: %w", err)
	}
	if err := progress.Validate(p); err != nil {
		return nil, err
	}
	return p, nil
}

func loadValidAppMap(root string) (*progress.AppMap, error) {
	appMap, err := progress.LoadAppMap(filepath.Join(root, "docs/content/building-goscrapling/architecture_plan/upstream-app-map.json"))
	if err != nil {
		return nil, fmt.Errorf("load app map: %w", err)
	}
	if err := progress.ValidateAppMap(appMap); err != nil {
		return nil, err
	}
	if err := progress.ValidateAppMapCoverage(root, appMap); err != nil {
		return nil, err
	}
	p, err := loadValid(root)
	if err != nil {
		return nil, err
	}
	if err := progress.ValidateAppMapReferences(root, appMap, p); err != nil {
		return nil, err
	}
	return appMap, nil
}

func writeAppMap(stdout io.Writer, root string) error {
	appMap, err := loadValidAppMap(root)
	if err != nil {
		return err
	}
	target := "docs/content/building-goscrapling/architecture_plan/upstream-app-map.md"
	if err := os.WriteFile(filepath.Join(root, target), []byte(progress.RenderAppMapMarkdown(appMap)), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	_, err = fmt.Fprintf(stdout, "app-map: regenerated %s\n", target)
	return err
}

func rewriteMarker(path, kind, body string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	output, err := progress.ReplaceMarker(string(input), kind, body)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
