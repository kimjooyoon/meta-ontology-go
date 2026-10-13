package bodycodegen

import "go/ast"

// Count resolved free calls in an already parsed body without changing its AST.
func countFreeBodyCalls(file *ast.File, name string) int {
	count := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok && id.Name == name && id.Obj == nil {
			count++
		}
		return true
	})
	return count
}
