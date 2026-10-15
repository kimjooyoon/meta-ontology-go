package bodycodegen

import "go/ast"

func (c *lengthIntegerContext) visit(node ast.Node) bool {
	switch item := node.(type) {
	case *ast.BinaryExpr:
		left, right := c.info.TypeOf(item.X), c.info.TypeOf(item.Y)
		item.X, item.Y = c.widen(item.X, right), c.widen(item.Y, left)
	case *ast.ReturnStmt:
		if len(item.Results) == 1 {
			item.Results[0] = c.widen(item.Results[0], c.output)
		}
	case *ast.AssignStmt:
		if len(item.Lhs) == 1 && len(item.Rhs) == 1 {
			item.Rhs[0] = c.widen(item.Rhs[0], c.info.TypeOf(item.Lhs[0]))
		}
	case *ast.ValueSpec:
		if item.Type != nil && len(item.Values) == 1 {
			item.Values[0] = c.widen(item.Values[0], c.info.TypeOf(item.Type))
		}
	case *ast.CallExpr:
		c.arguments(item)
	case *ast.KeyValueExpr:
		if key, ok := item.Key.(*ast.Ident); ok && c.info.Uses[key] != nil {
			item.Value = c.widen(item.Value, c.info.Uses[key].Type())
		}
	}
	return true
}
