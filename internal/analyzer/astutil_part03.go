package analyzer

import (
	"go/ast"
	"go/token"
)

func indexBase(expr ast.Expr) ast.Expr {
	switch current := expr.(type) {
	case *ast.IndexExpr:
		return current.X
	case *ast.IndexListExpr:
		return current.X
	default:
		return expr
	}
}
func expressionName(expr ast.Expr) string {
	switch current := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return current.Name
	case *ast.SelectorExpr:
		left := expressionName(current.X)
		if left == "" {
			return current.Sel.Name
		}
		return left + "." + current.Sel.Name
	default:
		return "<expression>"
	}
}
func spanFor(fileSet *token.FileSet, node ast.Node) Span {
	if node == nil {
		return Span{}
	}
	start := fileSet.PositionFor(node.Pos(), true)
	end := fileSet.PositionFor(node.End(), true)
	return Span{
		Filename: start.Filename,
		Start:    Position{Offset: start.Offset, Line: start.Line, Column: start.Column},
		End:      Position{Offset: end.Offset, Line: end.Line, Column: end.Column},
	}
}
