package main

import (
	"os"
	"path/filepath"
	"strings"
)

func outsideRoot(root, output string) (bool, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return false, err
	}
	return relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func writeExclusive(path string, payload []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
