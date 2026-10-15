package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
)

// Literal sequences use the same bounded, immutable scalar layout as repeated
// record fields. Each expression is evaluated once, in source order.
func validateSequenceLiteral(literal *ast.CompositeLit) error {
	sequence, ok := literal.Type.(*ast.ArrayType)
	if !ok || sequence.Len != nil || len(literal.Elts) > RepeatedFieldLimit {
		return fmt.Errorf("collection literal requires at most %d scalar elements", RepeatedFieldLimit)
	}
	element, ok := sequence.Elt.(*ast.Ident)
	if !ok || element.Name != "int64" && element.Name != "bool" && element.Name != "string" {
		return fmt.Errorf("collection literal requires []int64, []bool or []string")
	}
	for _, value := range literal.Elts {
		if _, keyed := value.(*ast.KeyValueExpr); keyed {
			return fmt.Errorf("collection literal requires elements in source order, without keyed indices")
		}
		if err := validateExpression(value); err != nil {
			return err
		}
	}
	return nil
}

func (e *integerBodyEvaluator) evaluateSequenceLiteral(literal *ast.CompositeLit, sequence *types.Slice) (any, error) {
	if err := validateSequenceLiteral(literal); err != nil {
		return nil, err
	}
	kind, optional, err := recordBodyScalarKind(sequence.Elem())
	if err != nil || optional || kind > 3 {
		return nil, fmt.Errorf("collection literal requires scalar elements")
	}
	result := &recordBodySequence{Kind: kind, Values: make([]recordBodyScalar, len(literal.Elts))}
	for i, element := range literal.Elts {
		value, err := e.evaluateExpression(element)
		if err != nil {
			return nil, err
		}
		result.Values[i], err = recordBodyScalarFromValue(value, kind)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
