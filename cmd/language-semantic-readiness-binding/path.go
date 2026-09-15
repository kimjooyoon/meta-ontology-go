package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func requireOutsideRepository(root, output string) error {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output must be outside repository: %s", outputPath)
	}
	return nil
}
