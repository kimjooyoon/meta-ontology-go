package syntax

import (
	"fmt"
	"sort"
	"strings"
)

func formatImports(imports []ImportDecl) ([]string, error) {
	paths := make([]string, len(imports))
	seen := make(map[string]bool, len(imports))
	for index, importDecl := range imports {
		if strings.TrimSpace(importDecl.Path) == "" {
			return nil, fmt.Errorf("package import path must be nonempty")
		}
		if seen[importDecl.Path] {
			return nil, fmt.Errorf("duplicate package import %q", importDecl.Path)
		}
		seen[importDecl.Path] = true
		paths[index] = importDecl.Path
	}
	sort.Strings(paths)
	return paths, nil
}
