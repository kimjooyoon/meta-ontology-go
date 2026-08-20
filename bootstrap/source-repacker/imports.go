package main

import (
	"go/ast"
	pathpkg "path"
	"strconv"
)

func importBindings(file *ast.File) (map[string]string, bool) {
	bindings := make(map[string]string)
	dot := false
	for _, spec := range file.Imports {
		name, path, ok := importBinding(spec)
		if !ok {
			continue
		}
		if name == "." {
			dot = true
			continue
		}
		bindings[name] = path
	}
	return bindings, dot
}

func importBinding(spec *ast.ImportSpec) (string, string, bool) {
	value, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return "", "", false
	}
	if spec.Name != nil {
		return spec.Name.Name, value, true
	}
	name := pathpkg.Base(value)
	if len(name) > 1 && name[0] == 'v' {
		if _, err := strconv.Atoi(name[1:]); err == nil {
			name = pathpkg.Base(pathpkg.Dir(value))
		}
	}
	return name, value, true
}

func selectorNames(declaration ast.Decl) map[string]bool {
	result := make(map[string]bool)
	ast.Inspect(declaration, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if identifier, ok := selector.X.(*ast.Ident); ok {
			result[identifier.Name] = true
		}
		return true
	})
	return result
}

func targetSupports(source, target parsedSource, declaration ast.Decl) bool {
	if source.DotImport {
		return false
	}
	for name := range selectorNames(declaration) {
		path, imported := source.Imports[name]
		if imported && target.Imports[name] != path {
			return false
		}
	}
	return true
}
