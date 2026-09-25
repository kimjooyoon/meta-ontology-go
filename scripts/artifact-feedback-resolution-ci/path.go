package main

import (
	"os"
	"path/filepath"
	"strings"
)

func outsideRoot(root, output string) (bool, error) {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false, err
	}
	outputPath, err := filepath.Abs(filepath.Clean(output))
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return false, err
	}
	return relative == ".." ||
		strings.HasPrefix(relative, ".."+string(os.PathSeparator)), nil
}
