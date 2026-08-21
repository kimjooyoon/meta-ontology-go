package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func canonicalOutputRoot(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("output root is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("absolute path: %w", err)
	}
	root := filepath.Clean(abs)
	if err := ensureOutputDirectory(root); err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("canonical path: %w", err)
	}
	return filepath.Clean(canonical), nil
}
