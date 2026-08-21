package analyzer

import (
	"go/ast"
	"slices"
)

func isDeclarationName(parent ast.Node, ident *ast.Ident) bool {
	switch current := parent.(type) {
	case *ast.AssignStmt:
		for _, left := range current.Lhs {
			if left == ident {
				return true
			}
		}
	case *ast.ValueSpec:
		if slices.Contains(current.Names, ident) {
			return true
		}
	case *ast.Field:
		if slices.Contains(current.Names, ident) {
			return true
		}
	}
	return false
}
