package bodycodegen

import (
	"fmt"
	"go/ast"
)

// CollectionIterationProductLimit bounds the product of nested sequence lengths.
// With the existing 256-item field limit, two collection levels fit this profile.
// This is an iteration-product bound, not a wall-time or instruction guarantee.
const CollectionIterationProductLimit = 65536

func validateCollectionNesting(loop *ast.RangeStmt) error {
	var inspect func(ast.Node, int) error
	inspect = func(root ast.Node, product int) error {
		var err error
		ast.Inspect(root, func(n ast.Node) bool {
			if err != nil {
				return false
			}
			switch nested := n.(type) {
			case *ast.RangeStmt:
				if product > CollectionIterationProductLimit/RepeatedFieldLimit {
					err = fmt.Errorf("nested collection iteration product exceeds%d", CollectionIterationProductLimit)
				} else {
					err = inspect(nested.Body, product*RepeatedFieldLimit)
				}
				return false
			case *ast.ForStmt:
				err = fmt.Errorf("nested counted loops retain their separate source profile")
				return false
			}
			return true
		})
		return err
	}
	return inspect(loop, 1)
}
