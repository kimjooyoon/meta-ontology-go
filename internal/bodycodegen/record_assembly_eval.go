package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// A value copies at most sixteen field string headers. Local whole-value
// assignments do not alias mutable maps or pointers to another record value.
type recordBodyValue struct {
	Type   string
	Count  int
	Values [16]string
}

func zeroRecordBodyValue(t types.Type) (recordBodyValue, error) {
	named, ok := t.(*types.Named)
	if !ok {
		return recordBodyValue{}, fmt.Errorf("record requires a nominal value type")
	}
	structure, ok := named.Underlying().(*types.Struct)
	if !ok || structure.NumFields() < 1 || structure.NumFields() > 16 {
		return recordBodyValue{}, fmt.Errorf("record layout exceeds value bounds")
	}
	for field := range structure.Fields() {
		if field.Type() != types.Typ[types.String] {
			return recordBodyValue{}, fmt.Errorf("record fields must be strings")
		}
	}
	return recordBodyValue{Type: named.Obj().Name(), Count: structure.NumFields()}, nil
}

func (e *integerBodyEvaluator) evaluateRecordLiteral(literal *ast.CompositeLit) (any, error) {
	value, err := zeroRecordBodyValue(e.information.Types[literal].Type)
	if err != nil {
		return nil, err
	}
	structure := e.information.Types[literal].Type.Underlying().(*types.Struct)
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("record evaluator requires named fields")
		}
		text, err := e.evaluateExpression(pair.Value)
		if err != nil {
			return nil, err
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("record key requires a name")
		}
		for i := 0; i < structure.NumFields(); i++ {
			if structure.Field(i).Name() == key.Name {
				field, ok := text.(string)
				if !ok {
					return nil, fmt.Errorf("record field requires text")
				}
				value.Values[i] = field
			}
		}
	}
	return value, nil
}

func (e *integerBodyEvaluator) evaluateRecordField(selector *ast.SelectorExpr) (any, error) {
	value, err := e.evaluateExpression(selector.X)
	if err != nil {
		return nil, err
	}
	record, ok := value.(recordBodyValue)
	if !ok {
		return nil, fmt.Errorf("field read requires a record value")
	}
	structure := e.information.Types[selector.X].Type.Underlying().(*types.Struct)
	for i := 0; i < structure.NumFields(); i++ {
		if structure.Field(i).Name() == selector.Sel.Name {
			return record.Values[i], nil
		}
	}
	return nil, fmt.Errorf("record field is not declared")
}

func evaluateRecordAssembly(ctx context.Context, source []byte, activity string, records []RecordType,
	cases []assemblyspec.ValueCase) ([]RecordAssemblyCase, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "record-evaluation.go", source, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	function, ok := findFunction(file, activity)
	if !ok {
		return nil, fmt.Errorf("record function is missing")
	}
	e := integerBodyEvaluator{context: ctx, information: types.Info{Types: make(map[ast.Expr]types.TypeAndValue),
		Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}}
	if _, err = new(types.Config).Check(file.Name.Name, fset, []*ast.File{file}, &e.information); err != nil {
		return nil, err
	}
	var inputs []types.Object
	for _, parameter := range function.Type.Params.List {
		for _, name := range parameter.Names {
			inputs = append(inputs, e.information.Defs[name])
		}
	}
	output := e.information.Types[function.Type.Results.List[0].Type].Type
	record := recordTypeByName(records, output.(*types.Named).Obj().Name())
	if record == nil {
		return nil, fmt.Errorf("record output contract is missing")
	}
	results := make([]RecordAssemblyCase, 0, len(cases))
	for _, c := range cases {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := e.evaluateRecordCase(function, inputs, output, *record, records, c)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, ctx.Err()
}

func (e *integerBodyEvaluator) evaluateRecordCase(function *ast.FuncDecl, inputs []types.Object, output types.Type,
	record RecordType, records []RecordType, c assemblyspec.ValueCase) (RecordAssemblyCase, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(c.Inputs), &raw); err != nil || len(raw) != len(inputs) {
		return RecordAssemblyCase{}, fmt.Errorf("value_case input must have exactly %d positional values", len(inputs))
	}
	e.environment = make(map[types.Object]any, len(inputs))
	for i, input := range inputs {
		value, err := decodeRecordCaseValue(raw[i], input.Type(), records)
		if err != nil {
			return RecordAssemblyCase{}, fmt.Errorf("value_case input %d: %w", i, err)
		}
		e.environment[input] = value
	}
	expected, err := decodeRecordCaseValue([]byte(c.Expected), output, records)
	if err != nil {
		return RecordAssemblyCase{}, fmt.Errorf("value_case expected: %w", err)
	}
	value, returned, err := e.evaluateBlock(function.Body)
	if err != nil || !returned {
		return RecordAssemblyCase{}, fmt.Errorf("record case did not return: %v", err)
	}
	actual, ok := value.(recordBodyValue)
	if !ok {
		return RecordAssemblyCase{}, fmt.Errorf("record case returned %T", value)
	}
	wanted := expected.(recordBodyValue)
	result := RecordAssemblyCase{Inputs: json.RawMessage(c.Inputs), Expected: json.RawMessage(c.Expected), Passed: actual == wanted}
	object := make(map[string]string, len(record.Fields))
	for i, field := range record.Fields {
		object[field.Name] = actual.Values[i]
		result.Fields = append(result.Fields, RecordAssemblyField{ID: field.ID, Name: field.Name, Expected: wanted.Values[i],
			Actual: actual.Values[i], Passed: actual.Values[i] == wanted.Values[i]})
	}
	result.Actual, _ = json.Marshal(object)
	return result, nil
}

func decodeRecordCaseValue(raw []byte, t types.Type, records []RecordType) (any, error) {
	if bytes.Equal(raw, []byte("null")) {
		return nil, fmt.Errorf("null is outside the pure value profile")
	}
	if named, ok := t.(*types.Named); ok {
		record := recordTypeByName(records, named.Obj().Name())
		if record == nil {
			return nil, fmt.Errorf("record type is not declared")
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil || len(object) != len(record.Fields) {
			return nil, fmt.Errorf("record requires exactly its declared fields")
		}
		value, err := zeroRecordBodyValue(t)
		if err != nil {
			return nil, err
		}
		for i, field := range record.Fields {
			fieldRaw, ok := object[field.Name]
			if !ok {
				return nil, fmt.Errorf("record field %q is missing", field.Name)
			}
			text, err := decodeRecordCaseValue(fieldRaw, types.Typ[types.String], records)
			if err != nil {
				return nil, err
			}
			value.Values[i] = text.(string)
		}
		return value, nil
	}
	switch t {
	case types.Typ[types.Int64]:
		return strconv.ParseInt(string(raw), 10, 64)
	case types.Typ[types.Bool]:
		if string(raw) == "true" {
			return true, nil
		}
		if string(raw) == "false" {
			return false, nil
		}
	case types.Typ[types.String]:
		var text string
		if err := json.Unmarshal(raw, &text); err == nil && len(text) <= 1024 && utf8.ValidString(text) {
			return text, nil
		}
	}
	return nil, fmt.Errorf("value does not match %s or exceeds its text bound", t)
}
