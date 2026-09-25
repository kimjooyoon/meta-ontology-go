package transformationeffect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func invokeSplitGoEvaluator(contractRaw, evidenceRaw []byte) (report any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("SplitGo evaluator invocation panicked: %v", recovered)
		}
	}()
	fn := reflect.ValueOf(operationconformance.Evaluate)
	typ := fn.Type()
	if typ.NumIn() != 2 || typ.NumOut() == 0 || typ.NumOut() > 2 {
		return nil, fmt.Errorf("unsupported SplitGo evaluator signature %s", typ)
	}
	contractValue, err := splitGoEvaluatorArgument(contractRaw, typ.In(0))
	if err != nil {
		return nil, fmt.Errorf("decode SplitGo contract argument: %w", err)
	}
	evidenceValue, err := splitGoEvaluatorArgument(evidenceRaw, typ.In(1))
	if err != nil {
		return nil, fmt.Errorf("decode SplitGo evidence argument: %w", err)
	}
	results := fn.Call([]reflect.Value{contractValue, evidenceValue})
	if len(results) == 2 {
		errorType := reflect.TypeFor[error]()
		if !results[1].Type().Implements(errorType) {
			return nil, fmt.Errorf("unsupported SplitGo evaluator error result %s", results[1].Type())
		}
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
	}
	return results[0].Interface(), nil
}

func splitGoEvaluatorArgument(raw []byte, target reflect.Type) (reflect.Value, error) {
	if target.Kind() == reflect.String {
		value := reflect.New(target).Elem()
		value.SetString(string(raw))
		return value, nil
	}
	if target.Kind() == reflect.Slice && target.Elem().Kind() == reflect.Uint8 {
		value := reflect.New(target).Elem()
		value.SetBytes(bytes.Clone(raw))
		return value, nil
	}
	if target.Kind() == reflect.Pointer {
		value := reflect.New(target.Elem())
		if err := json.Unmarshal(raw, value.Interface()); err != nil {
			return reflect.Value{}, err
		}
		return value, nil
	}
	value := reflect.New(target)
	if err := json.Unmarshal(raw, value.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return value.Elem(), nil
}
