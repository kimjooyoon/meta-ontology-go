package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
)

// Optional record references retain their storage origin through local aliases.
// The current value profile permits record copies and whole-field replacement;
// field mutation through a reference is outside its immutable value semantics.
func validateImmutableRecordReferences(file *ast.File, information *types.Info) error {
	var failure error
	ast.Inspect(file, func(node ast.Node) bool {
		write, ok := node.(*ast.AssignStmt)
		if !ok || failure != nil {
			return failure == nil
		}
		for _, target := range write.Lhs {
			field, ok := target.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			receiver := information.TypeOf(field.X)
			if receiver == nil {
				failure = fmt.Errorf("record field write %q requires a typed receiver", field.Sel.Name)
				return false
			}
			if _, pointer := types.Unalias(receiver).Underlying().(*types.Pointer); pointer {
				failure = fmt.Errorf("cannot mutate field %q through an immutable record reference", field.Sel.Name)
				return false
			}
		}
		return failure == nil
	})
	return failure
}
