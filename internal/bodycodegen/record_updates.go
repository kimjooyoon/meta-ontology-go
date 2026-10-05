package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
)

// Syntactic local validation precedes ordinary Go checking of the receiver's
// nominal record type, declared field and assigned scalar value.
func assignmentLocalName(target ast.Expr) (string, bool) {
	if selector, ok := target.(*ast.SelectorExpr); ok {
		target = ast.Unparen(selector.X)
	}
	name, ok := target.(*ast.Ident)
	if !ok {
		return "", false
	}
	return name.Name, true
}

func (e *integerBodyEvaluator) evaluateRecordUpdate(target *ast.SelectorExpr, expression ast.Expr) error {
	name, ok := ast.Unparen(target.X).(*ast.Ident)
	if !ok {
		return fmt.Errorf("field assignment requires a local record")
	}
	object := e.information.Uses[name]
	value, ok := e.environment[object].(recordBodyValue)
	if object == nil || !ok {
		return fmt.Errorf("field assignment has no record binding")
	}
	assigned, err := e.evaluateExpression(expression)
	if err != nil {
		return err
	}
	structure, ok := object.Type().Underlying().(*types.Struct)
	if !ok || structure.NumFields() != value.Count {
		return fmt.Errorf("field assignment has a different record layout")
	}
	for i := 0; i < value.Count; i++ {
		if structure.Field(i).Name() == target.Sel.Name {
			if structure.Field(i).Type() == types.Typ[types.Bool] {
				boolean, ok := assigned.(bool)
				if !ok {
					return fmt.Errorf("Boolean record field assignment requires a Boolean")
				}
				value.Values[i].Boolean = boolean
			} else {
				text, ok := assigned.(string)
				if !ok {
					return fmt.Errorf("text record field assignment requires text")
				}
				value.Values[i].Text = text
			}
			e.environment[object] = value
			return nil
		}
	}
	return fmt.Errorf("record field assignment is not declared")
}
