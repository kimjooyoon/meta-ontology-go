package main

import (
	"go/ast"
	"go/token"
	pathpkg "path"
	"strconv"
	"strings"
)

func importsFor(file *ast.File, declarations []ast.Decl) ([]ast.Decl, []*ast.ImportSpec) {
	used := selectorNames(declarations)
	importDecls := make([]ast.Decl, 0)
	imports := make([]*ast.ImportSpec, 0)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.IMPORT {
			continue
		}
		copyOf := *general
		copyOf.Specs = nil
		for _, raw := range general.Specs {
			spec := raw.(*ast.ImportSpec)
			if importRequired(spec, used) {
				copyOf.Specs = append(copyOf.Specs, spec)
				imports = append(imports, spec)
			}
		}
		if len(copyOf.Specs) != 0 {
			importDecls = append(importDecls, &copyOf)
		}
	}
	return importDecls, imports
}

func selectorNames(declarations []ast.Decl) map[string]bool {
	used := make(map[string]bool)
	for _, declaration := range declarations {
		ast.Inspect(declaration, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if identifier, ok := selector.X.(*ast.Ident); ok {
				used[identifier.Name] = true
			}
			return true
		})
	}
	return used
}

func importRequired(spec *ast.ImportSpec, used map[string]bool) bool {
	if spec.Name != nil {
		return spec.Name.Name == "_" || spec.Name.Name == "." || used[spec.Name.Name]
	}
	value, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return true
	}
	name := pathpkg.Base(value)
	if len(name) > 1 && name[0] == 'v' {
		if _, err := strconv.Atoi(name[1:]); err == nil {
			name = pathpkg.Base(pathpkg.Dir(value))
		}
	}
	return name == "C" || used[strings.TrimSpace(name)]
}
