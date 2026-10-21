package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// sourceModuleRoot chooses the nearest go.mod without changing repository-relative
// source and generated-file identities.
func sourceModuleRoot(root, logical string) (string, error) {
	path, err := safePath(root, logical)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return directoryModuleRoot(root, filepath.Dir(resolved))
}

func directoryModuleRoot(root, directory string) (string, error) {
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
