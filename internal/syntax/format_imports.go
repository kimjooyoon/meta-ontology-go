package syntax

import (
	"fmt"
	"sort"
	"strings"
)

func formatImports(imports []ImportDecl) ([]ImportDecl, error) {
	formatted := append([]ImportDecl(nil), imports...)
	seen := make(map[string]bool, len(imports))
	aliases := make(map[string]bool, len(imports))
	for _, importDecl := range imports {
		if strings.TrimSpace(importDecl.Path) == "" {
			return nil, fmt.Errorf("package import path must be nonempty")
		}
		if importDecl.Alias != "" {
			if err := validateIdentifier(importDecl.Alias, "package import alias"); err != nil {
				return nil, err
			}
		}
		if seen[importDecl.Path] {
			return nil, fmt.Errorf("duplicate package import %q", importDecl.Path)
		}
		seen[importDecl.Path] = true
		if importDecl.Alias != "" {
			if aliases[importDecl.Alias] {
				return nil, fmt.Errorf("duplicate package import alias %q", importDecl.Alias)
			}
			aliases[importDecl.Alias] = true
		}
	}
	sort.Slice(formatted, func(i, j int) bool { return formatted[i].Path < formatted[j].Path })
	return formatted, nil
}
