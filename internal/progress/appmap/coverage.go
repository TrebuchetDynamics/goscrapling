package appmap

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
