package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func secureSourcePath(root, subject string) (string, error) {
	if subject == "" || filepath.IsAbs(subject) {
		return "", fmt.Errorf("subject must be a relative file path")
	}
	clean := filepath.Clean(filepath.FromSlash(subject))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("subject escapes repository: %q", subject)
	}
	rootPath, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	target, err := filepath.EvalSymlinks(filepath.Join(rootPath, clean))
	if err != nil {
		return "", fmt.Errorf("resolve subject: %w", err)
	}
	relative, err := filepath.Rel(rootPath, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("subject escapes repository: %q", subject)
	}
	return target, nil
}
