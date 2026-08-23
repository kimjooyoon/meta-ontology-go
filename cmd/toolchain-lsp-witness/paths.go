package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func requireExternal(root string, paths ...string) error {
	rootPath, err := filepath.Abs(root)
	if err != nil { return err }
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil { return err }
		relative, err := filepath.Rel(rootPath, absolute)
		if err != nil { return err }
		if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			return fmt.Errorf("output/check path must be outside repository: %s", path)
		}
	}
	return nil
}
