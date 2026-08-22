package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateExternalPaths(value config) error {
	info, err := os.Stat(value.root)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("repository root is unavailable")
	}
	paths := []string{value.before, value.after}
	if value.output != "" {
		paths = append(paths, value.output)
	} else {
		paths = append(paths, value.check)
	}
	for _, candidate := range paths {
		if err := requireOutside(value.root, candidate); err != nil {
			return err
		}
	}
	return nil
}

func requireOutside(root, candidate string) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absoluteCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(absoluteRoot, absoluteCandidate)
	if err != nil {
		return err
	}
	parentPrefix := ".." + string(filepath.Separator)
	if relative == "." || (relative != ".." &&
		!strings.HasPrefix(relative, parentPrefix)) {
		return fmt.Errorf("artifact path must be outside repository: %s", candidate)
	}
	return nil
}
