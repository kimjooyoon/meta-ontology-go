package adkcheck

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sourcePath(root string, args CheckArgs) (string, error) {
	if args.Path == "" || filepath.IsAbs(args.Path) {
		return "", errors.New("path must be relative to the configured source root")
	}
	source, err := filepath.EvalSymlinks(filepath.Join(root, args.Path))
	if err != nil {
		return "", fmt.Errorf("resolve source path: %w", err)
	}
	rel, err := filepath.Rel(root, source)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("source path escapes the configured root")
	}
	info, err := os.Stat(source)
	if err != nil || !info.Mode().IsRegular() || filepath.Ext(source) != ".gooo" {
		return "", errors.New("path must name a regular .gooo file")
	}
	return rel, nil
}
