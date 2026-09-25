package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sourcePath(root, logical string) (string, error) {
	if logical == "" || filepath.IsAbs(logical) || strings.ContainsRune(logical, 0) {
		return "", fmt.Errorf("unsafe logical path %q", logical)
	}
	clean := filepath.Clean(filepath.FromSlash(logical))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe logical path %q", logical)
	}
	return filepath.Join(root, clean), nil
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func packageShapeSatisfied(root string, recipe packagePartitionRecipe) bool {
	return requirePackageShape(root, recipe) == nil
}

func requirePackageShape(root string, recipe packagePartitionRecipe) error {
	branch, err := sourcePath(root, recipe.Subject)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(branch)
	if err != nil || len(entries) != recipe.ExpectedShape.BranchEntries {
		return fmt.Errorf("partition branch shape mismatch")
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("partition branch mixes entry kinds")
		}
	}
	for leaf, expected := range recipe.ExpectedShape.Leaves {
		entries, err := os.ReadDir(branch + string(os.PathSeparator) + leaf)
		if err != nil || len(entries) != expected || len(entries) > recipe.ExpectedShape.MaxEntries {
			return fmt.Errorf("partition leaf shape mismatch: %s", leaf)
		}
	}
	return nil
}
