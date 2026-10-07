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

// Scalar slots keep records value-shaped and bounded; local copies never alias.
type recordBodyScalar struct {
	Text    string
	Boolean bool
	Integer int64
	Kind    uint8
	Present bool
}

type recordBodyOptionalScalar struct {
	Present bool
	Value   recordBodyScalar
}

type recordBodyValue struct {
	Type   string
	Count  int
	Values [16]recordBodyScalar
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
	value := recordBodyValue{Type: named.Obj().Name(), Count: structure.NumFields()}
	for field := 0; field < structure.NumFields(); field++ {
		kind, optional, err := recordBodyScalarKind(structure.Field(field).Type())
		if err != nil {
			return recordBodyValue{}, err
		}
		value.Values[field] = recordBodyScalar{Kind: kind, Present: !optional}
	}
	return value, nil
}

func recordBodyScalarKind(valueType types.Type) (uint8, bool, error) {
	optional := false
	if pointer, ok := valueType.(*types.Pointer); ok {
		optional = true
		valueType = pointer.Elem()
	}
	basic, ok := valueType.Underlying().(*types.Basic)
	if !ok {
		return 0, optional, fmt.Errorf("record field type is outside the scalar profile")
	}
	switch basic.Kind() {
	case types.String:
		return 1, optional, nil
	case types.Bool:
		return 2, optional, nil
	case types.Int64:
		return 3, optional, nil
	default:
		return 0, optional, fmt.Errorf("record field type is outside the scalar profile")
	}
}

func recordBodyScalarFromValue(value any, kind uint8) (recordBodyScalar, error) {
	scalar := recordBodyScalar{Kind: kind, Present: true}
	switch kind {
	case 1:
		text, ok := value.(string)
		if !ok {
			return recordBodyScalar{}, fmt.Errorf("record field requires text, got %T", value)
		}
		scalar.Text = text
	case 2:
		truth, ok := value.(bool)
		if !ok {
			return recordBodyScalar{}, fmt.Errorf("record field requires a Boolean, got %T", value)
		}
		scalar.Boolean = truth
	case 3:
		integer, ok := value.(int64)
		if !ok {
			return recordBodyScalar{}, fmt.Errorf("record field requires an integer, got %T", value)
		}
		scalar.Integer = integer
	default:
		return recordBodyScalar{}, fmt.Errorf("record field has an unsupported scalar kind")
	}
	return scalar, nil
}

func recordBodyScalarForField(value any, fieldType types.Type) (recordBodyScalar, error) {
	kind, optional, err := recordBodyScalarKind(fieldType)
	if err != nil {
		return recordBodyScalar{}, err
	}
	if optional {
		pointer, ok := value.(recordBodyOptionalScalar)
		if !ok {
			return recordBodyScalar{}, fmt.Errorf("optional record field assignment requires a pointer value or nil")
		}
		if !pointer.Present {
			return recordBodyScalar{Kind: kind}, nil
		}
		return recordBodyScalarFromValue(recordBodyScalarValue(pointer.Value), kind)
	}
	if _, ok := value.(recordBodyOptionalScalar); ok {
		return recordBodyScalar{}, fmt.Errorf("required record field assignment does not accept an optional value")
	}
	return recordBodyScalarFromValue(value, kind)
}

func recordBodyScalarValue(value recordBodyScalar) any {
	switch value.Kind {
	case 2:
		return value.Boolean
	case 3:
		return value.Integer
	default:
		return value.Text
	}
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
		assigned, err := e.evaluateExpression(pair.Value)
		if err != nil {
			return nil, err
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("record key requires a name")
		}
		for i := 0; i < structure.NumFields(); i++ {
			if structure.Field(i).Name() == key.Name {
				value.Values[i], err = recordBodyScalarForField(assigned, structure.Field(i).Type())
				if err != nil {
					return nil, err
				}
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
			if _, optional, err := recordBodyScalarKind(structure.Field(i).Type()); err != nil {
				return nil, err
			} else if optional {
				return recordBodyOptionalScalar{Present: record.Values[i].Present, Value: record.Values[i]}, nil
			}
			return recordBodyScalarValue(record.Values[i]), nil
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
	object := make(map[string]any, len(record.Fields))
	for i, field := range record.Fields {
		actualScalar, expectedScalar := actual.Values[i], wanted.Values[i]
		actualValue, expectedValue := "ABSENT", "ABSENT"
		if actualScalar.Present {
			actualValue = fmt.Sprint(recordScalarJSONValue(actualScalar))
		}
		if expectedScalar.Present {
			expectedValue = fmt.Sprint(recordScalarJSONValue(expectedScalar))
		}
		typeID := ""
		if field.TypeID == "urn:gooo:type:boolean" || field.TypeID == "urn:gooo:type:integer" {
			typeID = field.TypeID
		}
		fieldResult := RecordAssemblyField{ID: field.ID, Name: field.Name, TypeID: typeID, Expected: expectedValue,
			Actual: actualValue, Passed: actualScalar == expectedScalar}
		if field.Presence == "optional" {
			fieldResult.Presence = field.Presence
			fieldResult.ExpectedPresent = new(expectedScalar.Present)
			fieldResult.ActualPresent = new(actualScalar.Present)
		}
		result.Fields = append(result.Fields, fieldResult)
		if actualScalar.Present || field.Presence != "optional" {
			object[field.Name] = recordScalarJSONValue(actualScalar)
		}
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
		required := 0
		for _, field := range record.Fields {
			if field.Presence != "optional" {
				required++
			}
		}
		if err := json.Unmarshal(raw, &object); err != nil || len(object) < required || len(object) > len(record.Fields) {
			return nil, fmt.Errorf("record requires exactly its declared fields")
		}
		declared := make(map[string]bool, len(record.Fields))
		for _, field := range record.Fields {
			declared[field.Name] = true
		}
		for name := range object {
			if !declared[name] {
				return nil, fmt.Errorf("record field %q is not declared", name)
			}
		}
		value, err := zeroRecordBodyValue(t)
		if err != nil {
			return nil, err
		}
		for i, field := range record.Fields {
			fieldRaw, ok := object[field.Name]
			if !ok {
				if field.Presence == "optional" {
					continue
				}
				return nil, fmt.Errorf("record field %q is missing", field.Name)
			}
			fieldType := types.Type(types.Typ[types.String])
			if field.TypeID == "urn:gooo:type:boolean" {
				fieldType = types.Typ[types.Bool]
			} else if field.TypeID == "urn:gooo:type:integer" {
				fieldType = types.Typ[types.Int64]
			}
			decoded, err := decodeRecordCaseValue(fieldRaw, fieldType, records)
			if err != nil {
				return nil, err
			}
			kind, _, err := recordBodyScalarKind(fieldType)
			if err != nil {
				return nil, err
			}
			value.Values[i], err = recordBodyScalarFromValue(decoded, kind)
			if err != nil {
				return nil, err
			}
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

func recordScalarJSONValue(value recordBodyScalar) any {
	return recordBodyScalarValue(value)
}
