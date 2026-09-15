package main

import (
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
		if target == "" {
			continue
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
			return fmt.Errorf("proposal promotion output must be outside the repository root")
		}
	}
	return nil
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
