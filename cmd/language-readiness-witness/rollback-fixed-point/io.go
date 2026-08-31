package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func requireExternal(root, target string) error {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return err
	}
	targetPath, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(rootPath, targetPath)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("rollback output must be outside the repository root")
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
		return fmt.Errorf("rollback fixed-point replay mismatch")
	}
	return nil
}
