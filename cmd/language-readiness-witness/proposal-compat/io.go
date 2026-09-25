package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func requireExternal(root string, targets ...string) error {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return err
	}
	for _, target := range targets {
		targetPath, pathErr := filepath.Abs(filepath.Clean(target))
		if pathErr != nil {
			return pathErr
		}
		relative, pathErr := filepath.Rel(rootPath, targetPath)
		if pathErr != nil {
			return pathErr
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("compatibility output must be outside the repository root")
		}
	}
	return nil
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func compareFile(path string, expected []byte) error {
	actual, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("compatibility replay mismatch: %s", path)
	}
	return nil
}
