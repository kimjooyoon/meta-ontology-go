package bodycodegen

import (
	"go/ast"
	"go/token"
)

func normalizeCountedInitializers(file *ast.File) int {
	units := 0
	ast.Inspect(file, func(node ast.Node) bool {
		loop, ok := node.(*ast.ForStmt)
		if !ok {
			return true
		}
		initial, ok := loop.Init.(*ast.AssignStmt)
		if !ok || initial.Tok != token.DEFINE || len(initial.Rhs) != 1 || !loopIntegerLiteral(initial.Rhs[0], 0) {
			return true
		}
		if _, alreadyTyped := initial.Rhs[0].(*ast.CallExpr); alreadyTyped {
			return true
		}
		value := initial.Rhs[0]
		initial.Rhs[0] = &ast.CallExpr{Fun: &ast.Ident{Name: "int64", NamePos: value.Pos()}, Args: []ast.Expr{value}}
		units += 2
		return true
	})
	return units
}
