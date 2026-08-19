package linecaps

import (
	"fmt"
	"path/filepath"
	"strings"
)

func resolvePath(root, path string) (string, string, error) {
	if path == "" || path == "." {
		return "", "", fmt.Errorf("linecaps path must not be empty")
	}
	fullPath := filepath.FromSlash(path)
	if !filepath.IsAbs(fullPath) {
		return filepath.ToSlash(path), filepath.Join(root, fullPath), nil
	}
	relative, err := filepath.Rel(root, fullPath)
	if err != nil {
		return "", "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("linecaps path escapes root: %q", path)
	}
	return filepath.ToSlash(relative), fullPath, nil
}
