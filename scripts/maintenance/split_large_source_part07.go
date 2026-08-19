package main

import (
	"go/ast"
	"path/filepath"
	"strings"
)

func usedImportsForDecl(file *ast.File, decl ast.Decl) map[string]struct{} {
	used := make(map[string]struct{})
	knownImports := make(map[string]struct{}, len(file.Imports))
	for _, imp := range file.Imports {
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else if path := strings.Trim(imp.Path.Value, "\""); path != "" {
			alias = filepath.Base(path)
		}
		if alias == "" {
			continue
		}
		knownImports[alias] = struct{}{}
		if alias == "." || alias == "_" {
			used[alias] = struct{}{}
		}
	}
	if decl == nil {
		return used
	}
	ast.Inspect(decl, func(node ast.Node) bool {
		sel, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok {
			if _, ok := knownImports[id.Name]; ok {
				used[id.Name] = struct{}{}
			}
		}
		return true
	})
	return used
}
