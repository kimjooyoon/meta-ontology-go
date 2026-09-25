package extractor

import (
	"go/ast"
	"go/types"
)

// A terminal value copy may update fields only when the value itself is returned.
// Pointer traversal, indexed storage and rebinding the root remain unsupported.
func returnTailReturnedValueField(expression ast.Expr, object types.Object, statements []ast.Stmt, info *types.Info) bool {
	if !returnTailValueFieldWrite(expression, object, info) || len(statements) == 0 {
		return false
	}
	terminal, ok := statements[len(statements)-1].(*ast.ReturnStmt)
	if !ok {
		return false
	}
	for _, result := range terminal.Results {
		identifier, ok := ast.Unparen(result).(*ast.Ident)
		if ok && info.ObjectOf(identifier) == object {
			return true
		}
	}
	return false
}

// LookupSelection binds each field to the checked object and detects implicit
// pointer traversal through embedded fields. No name-only field allowlist is used.
func returnTailValueFieldWrite(expression ast.Expr, object types.Object, info *types.Info) bool {
	if info == nil || object == nil || object.Type() == nil {
		return false
	}
	if _, ok := object.(*types.Var); !ok {
		return false
	}
	if _, ok := object.Type().Underlying().(*types.Struct); !ok {
		return false
	}
	fieldSeen := false
	for {
		switch value := ast.Unparen(expression).(type) {
		case *ast.SelectorExpr:
			receiver := info.TypeOf(value.X)
			field, ok := info.Uses[value.Sel].(*types.Var)
			if receiver == nil || !ok {
				return false
			}
			if _, ok := receiver.Underlying().(*types.Struct); !ok {
				return false
			}
			selection, ok := types.LookupSelection(receiver, true, field.Pkg(), field.Name())
			if !ok || selection.Kind() != types.FieldVal || selection.Indirect() || selection.Obj() != field {
				return false
			}
			fieldSeen = true
			expression = value.X
		case *ast.Ident:
			return fieldSeen && info.ObjectOf(value) == object
		default:
			return false
		}
	}
}
