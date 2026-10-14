package bodyexecution

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func bridgeLeafMode(mode, expression string, reads int) string {
	if mode != "scalar_rhs" || reads != 0 {
		return mode
	}
	e, err := parser.ParseExpr(expression)
	if err == nil && bridgeConstantExpression(e) {
		return "literal"
	}
	return "unresolved_rhs"
}

// This is a syntactic observation on the retained typed expression. It neither
// evaluates a helper nor infers constant behavior from equal runtime values.
func bridgeConstantExpression(expression ast.Expr) bool {
	switch e := expression.(type) {
	case *ast.BasicLit:
		return e.Kind == token.INT || e.Kind == token.STRING
	case *ast.Ident:
		return e.Name == "true" || e.Name == "false"
	case *ast.ParenExpr:
		return bridgeConstantExpression(e.X)
	case *ast.UnaryExpr:
		return (e.Op == token.ADD || e.Op == token.SUB || e.Op == token.NOT) && bridgeConstantExpression(e.X)
	case *ast.BinaryExpr:
		return bridgeConstantExpression(e.X) && bridgeConstantExpression(e.Y)
	case *ast.CallExpr:
		name, ok := e.Fun.(*ast.Ident)
		return ok && (name.Name == "int64" || name.Name == "string" || name.Name == "bool") && len(e.Args) == 1 && bridgeConstantExpression(e.Args[0])
	}
	return false
}
