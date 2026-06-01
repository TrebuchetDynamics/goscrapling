package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TrebuchetDynamics/goscrapling/cmd/progress/appmap"
	"github.com/TrebuchetDynamics/goscrapling/cmd/progress/scorecard"
	"github.com/TrebuchetDynamics/goscrapling/cmd/progress/surfaces"
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
		p, err := loadValid(root)
		if err != nil {
			return err
		}
		return surfaces.Write(stdout, root, p)
	case "map-validate":
		p, err := loadValid(root)
		if err != nil {
			return err
		}
		appMap, err := appmap.LoadValid(root, p)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "app-map: validated %d entries\n", len(appMap.Entries))
		return err
	case "map-write":
		p, err := loadValid(root)
		if err != nil {
			return err
		}
		appMap, err := appmap.LoadValid(root, p)
		if err != nil {
			return err
		}
		return appmap.Write(stdout, root, appMap)
	case "scorecard":
		p, err := loadValid(root)
		if err != nil {
			return err
		}
		return scorecard.Write(stdout, root, p)
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
