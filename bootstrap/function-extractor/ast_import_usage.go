package main

import (
	"go/ast"
	"sort"
)

func genericImportSpecs(decls []ast.Decl, imports []astImport, includeBlank bool) (map[*ast.GenDecl][]*ast.ImportSpec, error) {
	used := map[string]bool{}
	for _, decl := range decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					used[ident.Name] = true
				}
			}
			return true
		})
	}
	result := map[*ast.GenDecl][]*ast.ImportSpec{}
	for _, item := range imports {
		if item.name == "." {
			return nil, extractionError("validate-ast-imports", "select-imports", "UNSUPPORTED_DOT_IMPORT", "KNOWN_CONTRADICTION", "report-contradiction", []string{item.path})
		}
		if item.name == "_" && !includeBlank {
			continue
		}
		if item.name == "_" || used[genericImportName(item)] {
			result[item.decl] = append(result[item.decl], item.spec)
		}
	}
	for _, specs := range result {
		sort.SliceStable(specs, func(i, j int) bool { return specs[i].Pos() < specs[j].Pos() })
	}
	return result, nil
}
