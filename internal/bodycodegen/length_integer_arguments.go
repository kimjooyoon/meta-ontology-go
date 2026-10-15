package bodycodegen

import (
	"go/ast"
	"go/types"
)

func (c *lengthIntegerContext) arguments(call *ast.CallExpr) {
	signature, ok := c.info.TypeOf(call.Fun).(*types.Signature)
	if !ok || signature.Variadic() || signature.Params().Len() != len(call.Args) {
		return
	}
	for i, argument := range call.Args {
		call.Args[i] = c.widen(argument, signature.Params().At(i).Type())
	}
}
