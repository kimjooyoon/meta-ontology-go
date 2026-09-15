package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ensureOutputDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return fmt.Errorf("inspect directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output root %q is a symbolic link", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("output root %q is not a directory", path)
	}
	return nil
}
func resolveOutputPath(root, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || filepath.Base(name) != name {
		return "", fmt.Errorf("output name %q escapes its root", name)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("absolute root: %w", err)
	}
	target := filepath.Clean(filepath.Join(absRoot, name))
	if !pathContained(absRoot, target) {
		return "", fmt.Errorf("output name %q escapes its root", name)
	}
	if err := validateOutputTarget(target); err != nil {
		return "", err
	}
	return target, nil
}
func pathContained(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(rel) || rel == "." || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func validateOutputTarget(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect output: %w", err)
	}
	return validateRegularFile(path, info)
}
