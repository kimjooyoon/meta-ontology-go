package main

import (
	"os"
	"path/filepath"
	"strings"
)

func outsideRoot(root, target string) (bool, error) {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false, err
	}
	targetPath, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootPath, targetPath)
	if err != nil {
		return false, err
	}
	return relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)), nil
}
