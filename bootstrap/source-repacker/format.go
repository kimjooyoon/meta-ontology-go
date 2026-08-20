package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

func formatPruned(source []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, err
	}
	used := make(map[string]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			if identifier, ok := selector.X.(*ast.Ident); ok {
				used[identifier.Name] = true
			}
		return true
	})
	declarations := make([]ast.Decl, 0, len(file.Decls))
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.IMPORT {
			declarations = append(declarations, declaration)
			continue
		}
		copyOf := *general
		copyOf.Specs = nil
		for _, raw := range general.Specs {
			spec := raw.(*ast.ImportSpec)
			name, _, valid := importBinding(spec)
			if !valid || name == "_" || name == "." || used[name] {
				copyOf.Specs = append(copyOf.Specs, spec)
			}
		}
		if len(copyOf.Specs) != 0 {
			declarations = append(declarations, &copyOf)
		}
	}
	file.Decls = declarations
	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}
