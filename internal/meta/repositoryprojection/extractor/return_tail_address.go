package extractor

import (
	"go/ast"
	"go/token"
	"go/types"
)

// Addressing a typed composite literal allocates new storage. Its elements are
// still visited by the caller, including any addresses of existing bindings.
// Keep this separate from assignment lookup: literals are not assignable slots.
func returnTailAddressedObject(address *ast.UnaryExpr, info *types.Info) (types.Object, bool) {
	if address == nil || address.Op != token.AND || info == nil {
		return nil, false
	}
	expression := ast.Unparen(address.X)
	if _, literal := expression.(*ast.CompositeLit); !literal {
		return returnTailAssignedObject(expression, info)
	}
	valueType := info.TypeOf(expression)
	pointer, ok := info.TypeOf(address).(*types.Pointer)
	if valueType == nil || !ok || !types.Identical(pointer.Elem(), valueType) {
		return nil, false
	}
	switch valueType.Underlying().(type) {
	case *types.Struct, *types.Array, *types.Slice, *types.Map:
		return nil, true
	default:
		return nil, false
	}
}
