package appmap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress"
)

const markdownPath = "docs/content/building-goscrapling/architecture_plan/upstream-app-map.md"

// LoadValid loads the upstream app map and validates local coverage and progress references.
func LoadValid(root string, p *progress.Progress) (*progress.AppMap, error) {
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
	if err := progress.ValidateAppMapReferences(root, appMap, p); err != nil {
		return nil, err
	}
	return appMap, nil
}

// Write renders the validated upstream app map markdown surface.
func Write(stdout io.Writer, root string, appMap *progress.AppMap) error {
	if err := os.WriteFile(filepath.Join(root, markdownPath), []byte(progress.RenderAppMapMarkdown(appMap)), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", markdownPath, err)
	}
	_, err := fmt.Fprintf(stdout, "app-map: regenerated %s\n", markdownPath)
	return err
}
