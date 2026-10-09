package bodycodegen

import (
	"go/ast"
	"go/token"
)

// Preserve already accepted negative literals and their arena indices. Other
// integer negations use the existing subtract operation; native scalar typing
// rejects boolean operands before the recipe can be used.
func normalizeArithmeticExpression(node ast.Expr) (ast.Expr, bool) {
	expression, ok := node.(*ast.UnaryExpr)
	if !ok || expression.Op != token.SUB {
		return node, false
	}
	if literal, ok := expression.X.(*ast.BasicLit); ok && literal.Kind == token.INT {
		return node, false
	}
	return &ast.BinaryExpr{X: &ast.BasicLit{Kind: token.INT, Value: "0", ValuePos: expression.OpPos},
		Op: token.SUB, OpPos: expression.OpPos, Y: expression.X}, true
}
