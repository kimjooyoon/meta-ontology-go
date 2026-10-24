// Package moduleboundary finds the Go module that owns a repository source.
package moduleboundary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Directory returns the nearest go.mod's directory. When no manifest exists
// within the repository, the repository root remains the legacy fallback.
func Directory(root, directory string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	directory, err = filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	resolvedDir, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", err
	}
	within, err := filepath.Rel(resolvedRoot, resolvedDir)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("module directory escapes repository: %s", directory)
	}
	for current := resolvedDir; ; current = filepath.Dir(current) {
		info, statErr := os.Lstat(filepath.Join(current, "go.mod"))
		if statErr == nil {
			if !info.Mode().IsRegular() {
				return "", fmt.Errorf("module manifest is not a regular file: %s", current)
			}
			return current, nil
		}
		if statErr != nil && !os.IsNotExist(statErr) {
			return "", statErr
		}
		if current == resolvedRoot {
			return resolvedRoot, nil
		}
	}
}
