package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func safeLogical(name string) error {
	if name == "" || filepath.IsAbs(name) || strings.ContainsRune(name, 0) {
		return fmt.Errorf("unsafe logical path %q", name)
	}
	clean := filepath.ToSlash(filepath.Clean(name))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("unsafe logical path %q", name)
	}
	return nil
}

func prepareWork(root, work string) error {
	if work == "" || work == string(filepath.Separator) {
		return fmt.Errorf("safe empty work directory is required")
	}
	relative, err := filepath.Rel(root, work)
	if err != nil {
		return err
	}
	if relative == "." || !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("work directory must be outside the repository")
	}
	if _, err := os.Stat(work); !os.IsNotExist(err) {
		return fmt.Errorf("work directory must not exist: %s", work)
	}
	return os.MkdirAll(work, 0o755)
}

func languageFor(name string) string {
	switch filepath.Ext(name) {
	case ".go":
		return "go"
	case ".gooo":
		return "gooo"
	default:
		return ""
	}
}

func lineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func retainedBacking(logical string) (string, bool) {
	if logical == "go.mod" || logical == "go.sum" {
		return filepath.ToSlash(filepath.Join("module", logical)), true
	}
	if strings.HasPrefix(logical, "bootstrap/") ||
		strings.HasPrefix(logical, ".github/workflows/") {
		return logical, true
	}
	return "", false
}
