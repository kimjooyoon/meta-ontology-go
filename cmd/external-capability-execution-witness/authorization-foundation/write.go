package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeOutside(path string, raw []byte) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	target, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(cwd, target)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output must remain outside repository: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, raw, 0o644)
}
