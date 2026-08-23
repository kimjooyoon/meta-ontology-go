package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func requireExternal(root string, paths ...string) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(absoluteRoot, absolutePath)
		if err != nil {
			return err
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("path must be outside repository: %s", path)
		}
	}
	return nil
}
