package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolve(root, value string) string {
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(root, filepath.FromSlash(value))
}

func writeOutsideRepository(root, output string, raw []byte) error {
	target, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("receipt output must be outside the source repository")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, raw, 0o644)
}
