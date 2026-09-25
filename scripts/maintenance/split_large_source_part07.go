package main

import (
	"go/ast"
)

func usedImportsForDecl(file *ast.File, decl ast.Decl) map[string]struct{} {
	used := make(map[string]struct{})
	knownImports := make(map[string][]string, len(file.Imports))
	for _, spec := range buildImportMap(file) {
		key := spec.key()
		knownImports[spec.name] = append(knownImports[spec.name], key)
		if spec.name == "." || spec.name == "_" {
			used[key] = struct{}{}
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
			for _, key := range knownImports[id.Name] {
				used[key] = struct{}{}
			}
		}
		return true
	})
	return used
}

func (i importSpec) key() string {
	return i.name + "\x00" + i.path
}
