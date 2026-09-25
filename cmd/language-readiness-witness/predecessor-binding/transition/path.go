package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validatePaths(value config) error {
	for _, path := range []string{value.before, value.after, value.output + value.check} {
		root, err := filepath.Abs(value.root)
		if err != nil {
			return err
		}
		candidate, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, candidate)
		if err != nil {
			return err
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("binding artifact path must be outside repository")
		}
	}
	return nil
}
