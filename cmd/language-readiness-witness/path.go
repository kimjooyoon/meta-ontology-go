package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func requireExternal(root string, paths ...string) error {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return err
	}
	for _, path := range paths {
		target, err := filepath.Abs(filepath.Clean(path))
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(rootPath, target)
		if err != nil {
			return err
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("readiness witness paths must be outside the repository root")
		}
	}
	return nil
}
