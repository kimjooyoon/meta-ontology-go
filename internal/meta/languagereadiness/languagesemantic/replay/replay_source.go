package replay

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

func sourcePath(root, relativePath string) (string, string, error) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relativePath)))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("source path %q is outside the repository", relativePath)
	}
	if filepath.Ext(clean) != ".gooo" {
		return "", "", fmt.Errorf("source path %q is not a .gooo file", relativePath)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", "", err
	}
	target := filepath.Join(rootAbs, clean)
	relative, err := filepath.Rel(rootAbs, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("source path %q escapes the repository", relativePath)
	}
	return target, filepath.ToSlash(clean), nil
}

func physicalLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}
