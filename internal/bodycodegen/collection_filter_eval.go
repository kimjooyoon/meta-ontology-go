package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
)

func (e *integerBodyEvaluator) evaluateCollectionFilter(call *ast.CallExpr, input ast.Expr, predicate *ast.Ident) (any, error) {
	raw, err := e.evaluateExpression(input)
	if err != nil {
		return nil, err
	}
	values, ok := raw.(*recordBodySequence)
	if !ok || values == nil || len(values.Values) > RepeatedFieldLimit {
		return nil, fmt.Errorf("filter_values requires a collection of at most %d elements", RepeatedFieldLimit)
	}
	function := e.functions[e.information.Uses[predicate]]
	if function == nil || len(function.Type.Params.List) != 1 {
		return nil, fmt.Errorf("filter_values predicate has no unary typed function")
	}
	sequence, ok := e.information.Types[call].Type.(*types.Slice)
	if !ok {
		return nil, fmt.Errorf("filter_values output is not a scalar collection")
	}
	kind, optional, err := recordBodyScalarKind(sequence.Elem())
	if err != nil || optional || kind > 3 || kind != values.Kind {
		return nil, fmt.Errorf("filter_values output differs from its input scalar type")
	}
	result := &recordBodySequence{Kind: kind, Values: make([]recordBodyScalar, 0, len(values.Values))}
	for _, item := range values.Values {
		selected, err := e.invokePureFunction(function, func(frame *integerBodyEvaluator) error {
			parameter := e.information.Defs[function.Type.Params.List[0].Names[0]]
			value, err := coerceBodyValue(recordBodyScalarValue(item), parameter.Type())
			frame.environment[parameter] = value
			return err
		})
		if err != nil {
			return nil, err
		}
		keep, ok := selected.(bool)
		if !ok {
			return nil, fmt.Errorf("filter_values predicate did not return Boolean")
		}
		if keep {
			result.Values = append(result.Values, item)
		}
	}
	return result, nil
}
