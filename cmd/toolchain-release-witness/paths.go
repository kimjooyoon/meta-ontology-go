package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func requireExternal(root string, paths ...string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	for _, path := range paths {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." || (!strings.HasPrefix(relative, ".."+string(filepath.Separator)) && relative != "..") {
			return fmt.Errorf("output or receipt path must be outside repository: %s", path)
		}
	}
	return nil
}
