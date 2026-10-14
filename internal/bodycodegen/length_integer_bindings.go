package bodycodegen

import (
	"go/ast"
	"go/types"
)

func runtimeLengthBindings(file *ast.File, info *types.Info) map[types.Object]bool {
	lengths := make(map[types.Object]bool)
	for {
		before := len(lengths)
		ast.Inspect(file, func(node ast.Node) bool {
			var names []*ast.Ident
			var values []ast.Expr
			switch item := node.(type) {
			case *ast.ValueSpec:
				names, values = item.Names, item.Values
			case *ast.AssignStmt:
				for _, expression := range item.Lhs {
					name, _ := expression.(*ast.Ident)
					names = append(names, name)
				}
				values = item.Rhs
			}
			if len(names) != 1 || names[0] == nil || len(values) != 1 ||
				!dependsOnRuntimeLength(values[0], info, lengths) {
				return true
			}
			object := info.Defs[names[0]]
			if object == nil {
				object = info.Uses[names[0]]
			}
			if object != nil {
				lengths[object] = true
			}
			return true
		})
		if len(lengths) == before {
			return lengths
		}
	}
}
