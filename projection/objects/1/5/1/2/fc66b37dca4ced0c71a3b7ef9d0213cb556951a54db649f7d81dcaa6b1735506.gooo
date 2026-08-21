package linecaps

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func resolvePath(root, path string) (string, string, error) {
	if path == "" || path == "." {
		return "", "", fmt.Errorf("linecaps path must not be empty")
	}
	fullPath := filepath.FromSlash(path)
	if !filepath.IsAbs(fullPath) {
		return filepath.ToSlash(path), filepath.Join(root, fullPath), nil
	}
	relative, err := filepath.Rel(root, fullPath)
	if err != nil {
		return "", "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("linecaps path escapes root: %q", path)
	}
	return filepath.ToSlash(relative), fullPath, nil
}

func orderedMetricDirectories(directories map[string]*directoryNode) []string {
	entries := make([]string, 0, len(directories))
	for path := range directories {
		entries = append(entries, path)
	}
	sort.Slice(entries, func(i, j int) bool {
		iDepth, jDepth := directoryDepth(entries[i]), directoryDepth(entries[j])
		if iDepth != jDepth {
			return iDepth > jDepth
		}
		if entries[i] == "." {
			return false
		}
		if entries[j] == "." {
			return true
		}
		return entries[i] < entries[j]
	})
	return entries
}
