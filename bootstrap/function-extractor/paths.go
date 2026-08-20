package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

func extractionPath(root, logical string) (string, error) {
	if logical == "" || filepath.IsAbs(logical) || strings.ContainsRune(logical, 0) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	clean := filepath.Clean(filepath.FromSlash(logical))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	return filepath.Join(root, clean), nil
}

func extractionLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func editText(lines []string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}
