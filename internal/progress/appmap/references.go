package appmap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/progress/model"
)

func ValidateAppMapReferences(repoRoot string, m *AppMap, p *model.Progress) error {
	progressRows := make(map[string]struct{})
	for _, row := range model.Rows(p) {
		if name := strings.TrimSpace(row.Item.Name); name != "" {
			progressRows[name] = struct{}{}
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
