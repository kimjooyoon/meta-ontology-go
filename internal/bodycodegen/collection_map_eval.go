package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
)

func (e *integerBodyEvaluator) evaluateCollectionMap(call *ast.CallExpr, input ast.Expr, mapper *ast.Ident) (any, error) {
	raw, err := e.evaluateExpression(input)
	if err != nil {
		return nil, err
	}
	values, ok := raw.(*recordBodySequence)
	if !ok || values == nil || len(values.Values) > RepeatedFieldLimit {
		return nil, fmt.Errorf("map_values requires a collection of at most %d elements", RepeatedFieldLimit)
	}
	function := e.functions[e.information.Uses[mapper]]
	if function == nil || len(function.Type.Params.List) != 1 {
		return nil, fmt.Errorf("map_values mapper has no unary typed function")
	}
	sequence, ok := e.information.Types[call].Type.(*types.Slice)
	if !ok {
		return nil, fmt.Errorf("map_values output is not a scalar collection")
	}
	kind, optional, err := recordBodyScalarKind(sequence.Elem())
	if err != nil || optional || kind > 3 {
		return nil, fmt.Errorf("map_values output is outside its scalar profile")
	}
	result := &recordBodySequence{Kind: kind, Values: make([]recordBodyScalar, len(values.Values))}
	for i, item := range values.Values {
		value, err := e.invokePureFunction(function, func(frame *integerBodyEvaluator) error {
			parameter := e.information.Defs[function.Type.Params.List[0].Names[0]]
			value, err := coerceBodyValue(recordBodyScalarValue(item), parameter.Type())
			frame.environment[parameter] = value
			return err
		})
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
