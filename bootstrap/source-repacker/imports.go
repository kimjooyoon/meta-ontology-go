package main

import (
	"go/ast"
	pathpkg "path"
	"sort"
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

func requiredImports(source, target parsedSource, declaration ast.Decl) ([]importAddition, bool) {
	if source.DotImport {
		return nil, false
	}
	additions := make([]importAddition, 0)
	for name := range selectorNames(declaration) {
		path, imported := source.Imports[name]
		if !imported {
			continue
		}
		if current, exists := target.Imports[name]; exists {
			if current != path {
				return nil, false
			}
			continue
		}
		if name == "C" {
			return nil, false
		}
		additions = append(additions, importAddition{Name: name, Path: path})
	}
	sort.Slice(additions, func(i, j int) bool { return additions[i].Name < additions[j].Name })
	return additions, true
}
