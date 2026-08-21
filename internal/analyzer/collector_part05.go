package analyzer

import (
	"go/ast"
)

func addNames(blocked map[string]bool, names []*ast.Ident) {
	for _, name := range names {
		if name.Name != "_" {
			blocked[name.Name] = true
		}
	}
}
func addExprNames(blocked map[string]bool, expressions []ast.Expr) {
	for _, expression := range expressions {
		addExprName(blocked, expression)
	}
}
func addExprName(blocked map[string]bool, expression ast.Expr) {
	ident, ok := expression.(*ast.Ident)
	if ok && ident.Name != "_" {
		blocked[ident.Name] = true
	}
}
