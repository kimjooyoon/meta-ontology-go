package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func missingDirectoryChain(raw string) []string {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return nil
	}
	current := filepath.Clean(abs)
	var missing []string
	for {
		if _, err := os.Lstat(current); err == nil {
			break
		} else if !errors.Is(err, fs.ErrNotExist) {
			break
		}
		missing = append(missing, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return missing
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
	}
	return result
}

func cleanupGenerateDirectories(paths []string) error {
	var firstErr error
	ordered := append([]string(nil), paths...)
	sort.SliceStable(ordered, func(left, right int) bool { return len(ordered[left]) > len(ordered[right]) })
	for _, path := range ordered {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) && firstErr == nil {
			firstErr = fmt.Errorf("remove temporary output directory %q: %w", path, err)
		}
	}
	return firstErr
}
