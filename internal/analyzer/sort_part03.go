package analyzer

import (
	"go/ast"
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
		for _, name := range current.Names {
			if name == ident {
				return true
			}
		}
	case *ast.Field:
		for _, name := range current.Names {
			if name == ident {
				return true
			}
		}
	}
	return false
}
