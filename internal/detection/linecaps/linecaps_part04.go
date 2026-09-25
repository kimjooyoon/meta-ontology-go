package linecaps

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

func functionSpan(fset *token.FileSet, node ast.Node) (name string, start int, end int, ok bool) {
	switch function := node.(type) {
	case *ast.FuncDecl:
		name = function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
		return name, fset.Position(node.Pos()).Line, fset.Position(node.End()).Line, true
	case *ast.FuncLit:
		return "function literal", fset.Position(node.Pos()).Line, fset.Position(node.End()).Line, true
	default:
		return "", 0, 0, false
	}
}
func normalizePaths(root string, paths []string) ([]string, error) {
	result := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		normalized, err := normalizePath(root, path)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result, nil
}
func normalizePath(root, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("linecaps path must not be empty")
	}
	normalized := strings.ReplaceAll(path, "\\", "/")
	cleaned := filepath.Clean(filepath.FromSlash(normalized))
	if filepath.IsAbs(cleaned) {
		relative, err := filepath.Rel(root, cleaned)
		if err != nil {
			return "", err
		}
		if relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("linecaps path escapes root: %q", path)
		}
		return filepath.ToSlash(relative), nil
	}
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("linecaps path escapes root: %q", path)
	}
	return filepath.ToSlash(cleaned), nil
}
