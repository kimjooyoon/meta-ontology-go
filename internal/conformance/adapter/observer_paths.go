package adapter

import (
	"fmt"
	"path/filepath"
	"strings"
)

func canonicalObserverPath(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

func validateObserverPathDisjointness(paths ObserverPaths) error {
	pairs := []struct {
		name  string
		left  string
		right string
	}{
		{name: "source/output", left: paths.SourcePath, right: paths.OutputPath},
		{name: "source/temp", left: paths.SourcePath, right: paths.TempRoot},
		{name: "output/temp", left: paths.OutputPath, right: paths.TempRoot},
	}
	for _, pair := range pairs {
		if observerPathsOverlap(pair.left, pair.right) {
			return fmt.Errorf("observer %s paths overlap: %q and %q", pair.name, pair.left, pair.right)
		}
	}
	return nil
}

func observerPathsOverlap(left, right string) bool {
	return observerPathContains(left, right) || observerPathContains(right, left)
}

func observerPathContains(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return true
	}
	if relative == "." {
		return true
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
