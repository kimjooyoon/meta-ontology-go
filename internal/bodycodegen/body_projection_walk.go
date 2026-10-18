package bodycodegen

import "go/ast"

// InspectBodyProjection visits source operations, including the original input
// of a lowered map, without treating its implementation loop as authored code.
// Recognition checks the complete lowering shape, not reserved-looking names.
func InspectBodyProjection(root ast.Node, visit func(ast.Node) bool) {
	ast.Inspect(root, func(node ast.Node) bool {
		if !visit(node) {
			return false
		}
		if call, ok := node.(*ast.CallExpr); ok {
			if input, _, mapped := CollectionMapProjection(call); mapped {
				InspectBodyProjection(input, visit)
				visit(nil)
				return false
			}
			if input, _, filtered := CollectionFilterProjection(call); filtered {
				InspectBodyProjection(input, visit)
				visit(nil)
				return false
			}
		}
		return true
	})
}
