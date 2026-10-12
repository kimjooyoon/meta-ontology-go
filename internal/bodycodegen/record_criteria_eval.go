package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

// This evaluator consumes an already projected pure body and raw positional
// inputs. It produces a value without inventing an expected record for it.
type criteriaTupleEvaluator struct {
	evaluator integerBodyEvaluator
	function  *ast.FuncDecl
	inputs    []types.Object
	output    types.Type
	records   []RecordType
}

func newCriteriaTupleEvaluator(ctx context.Context, source []byte, activity string,
	records []RecordType, calls *localCallRecorder) (*criteriaTupleEvaluator, error) {
	if ctx == nil {
		return nil, fmt.Errorf("criteria evaluation requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "criteria-evaluation.go", source, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	function, ok := findFunction(file, activity)
	if !ok || function.Type.Params == nil || function.Type.Results == nil || len(function.Type.Results.List) != 1 {
		return nil, fmt.Errorf("criteria evaluation requires a typed single-result function")
	}
	information := types.Info{Types: make(map[ast.Expr]types.TypeAndValue),
		Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}
	if _, err := new(types.Config).Check(file.Name.Name, fset, []*ast.File{file}, &information); err != nil {
		return nil, err
	}
	p := &criteriaTupleEvaluator{function: function, records: records,
		output: information.Types[function.Type.Results.List[0].Type].Type,
		evaluator: integerBodyEvaluator{context: ctx, information: information,
			functions: pureEvaluatorFunctions(file, information), localCalls: calls}}
	for _, field := range function.Type.Params.List {
		for _, name := range field.Names {
			p.inputs = append(p.inputs, information.Defs[name])
		}
	}
	if len(p.inputs) < 1 || len(p.inputs) > 16 {
		return nil, fmt.Errorf("criteria evaluation requires 1..16 typed inputs")
	}
	return p, ctx.Err()
}

func (p *criteriaTupleEvaluator) bind(input string) error {
	if err := p.evaluator.context.Err(); err != nil {
		return err
	}
	var values []json.RawMessage
	if err := json.Unmarshal([]byte(input), &values); err != nil || len(values) != len(p.inputs) {
		return fmt.Errorf("criteria input needs exactly %d positional values", len(p.inputs))
	}
	frame := make(map[types.Object]any, len(p.inputs))
	for i, parameter := range p.inputs {
		value, err := decodeRecordCaseValue(values[i], parameter.Type(), p.records)
		if err != nil {
			return fmt.Errorf("criteria input %d: %w", i, err)
		}
		frame[parameter] = value
	}
	p.evaluator.environment, p.evaluator.callCount = frame, nil
	return nil
}

func (p *criteriaTupleEvaluator) evaluate(input string, caseIndex int) (json.RawMessage, error) {
	if err := p.bind(input); err != nil {
		return nil, err
	}
	e := &p.evaluator
	e.caseIndex = caseIndex
	if e.localCalls != nil {
		e.localCalls.observation.CasesStarted++
	}
	value, returned, err := e.evaluateBlock(p.function.Body)
	if err != nil {
		return nil, fmt.Errorf("criteria program did not return: %w", err)
	}
	if !returned {
		return nil, fmt.Errorf("criteria program did not return")
	}
	if err := e.context.Err(); err != nil {
		return nil, err
	}
	encoded, err := encodeCriteriaOutput(value, p.output, p.records)
	if err == nil && e.localCalls != nil {
		e.localCalls.observation.CasesCompleted++
	}
	return encoded, err
}

func encodeCriteriaOutput(value any, output types.Type, records []RecordType) (json.RawMessage, error) {
	coerced, err := coerceBodyValue(value, output)
	if err != nil {
		return nil, err
	}
	if output == types.Typ[types.Bool] || output == types.Typ[types.Int64] {
		return json.Marshal(coerced)
	}
	named, ok := output.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("criteria output must be Integer, Boolean or a declared record")
	}
	record := recordTypeByName(records, named.Obj().Name())
	actual, ok := coerced.(recordBodyValue)
	if record == nil || !ok {
		return nil, fmt.Errorf("criteria record result is outside its declared type")
	}
	object := make(map[string]any, len(record.Fields))
	for i, field := range record.Fields {
		value := actual.Values[i]
		if value.Present || field.Presence != "optional" {
			object[field.Name] = recordScalarJSONValue(value)
		}
	}
	return json.Marshal(object)
}
