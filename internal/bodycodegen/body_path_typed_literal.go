package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
)

// The SDK emits explicit int64 constants. Accept only that literal spelling;
// arbitrary conversions and calls do not belong to the closed typed arena.
func typedIntegerLiteral(call *ast.CallExpr) (ast.Expr, error) {
	name, ok := call.Fun.(*ast.Ident)
	if ok && name.Name == "int64" && len(call.Args) == 1 && !call.Ellipsis.IsValid() {
		expression := call.Args[0]
		literal := expression
		if negative, ok := literal.(*ast.UnaryExpr); ok && negative.Op == token.SUB {
			literal = negative.X
		}
		if value, ok := literal.(*ast.BasicLit); ok && value.Kind == token.INT {
			return expression, nil
		}
	}
	return nil, fmt.Errorf("recipe call must be an explicit int64 integer literal")
}
