package bodycodegen

import "go/ast"

// Without a syntactic len call there is no length-producing binding to widen.
// This avoids repeating a typecheck for scalar candidate programs that cannot
// use the length-context extension. A source-declared len remains eligible for
// the ordinary typed resolution; strings and comments are not calls.
func bodyMayUseLength(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "len" {
				found = true
			}
		}
		return !found
	})
	return found
}
