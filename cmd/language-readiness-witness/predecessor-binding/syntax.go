package main

import (
	"go/ast"
	"go/token"
)

func constantExpressions(file *ast.File) map[string]ast.Expr {
	result := make(map[string]ast.Expr)
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, specification := range group.Specs {
			value, ok := specification.(*ast.ValueSpec)
			if !ok || len(value.Values) != len(value.Names) {
				continue
			}
			for index, name := range value.Names {
				result[name.Name] = value.Values[index]
			}
		}
	}
	return result
}

func parameterNames(function *ast.FuncDecl) map[string]bool {
	result := make(map[string]bool)
	if function.Type.Params == nil {
		return result
	}
	for _, field := range function.Type.Params.List {
		for _, name := range field.Names {
			result[name.Name] = true
		}
	}
	return result
}
