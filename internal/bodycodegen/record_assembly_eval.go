package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// Pointer and scalar fields occupy the old cell's alignment padding. Adding a
// nominal child reference does not enlarge each of the 32 value slots.
type recordBodyScalar struct {
	Text       string
	Integer    int64
	Collection *recordBodySequence
	Record     *recordBodyValue
	Kind       uint8
	Present    bool
	Boolean    bool
}

type recordBodyOptionalScalar struct {
	Present bool
	Value   recordBodyScalar
}

type recordBodyValue struct {
	Type   string
	Count  int
	Values [RecordFieldLimit]recordBodyScalar
	layout *types.Struct
	height int
}

func zeroRecordBodyValue(t types.Type) (recordBodyValue, error) {
	return newRecordBodyZero(t)
}

func recordBodyScalarKind(valueType types.Type) (uint8, bool, error) {
	optional := false
	if pointer, ok := valueType.(*types.Pointer); ok {
		optional = true
		valueType = pointer.Elem()
	}
	if sequence, ok := valueType.Underlying().(*types.Slice); ok {
		kind, nestedOptional, err := recordBodyScalarKind(sequence.Elem())
		if err != nil || nestedOptional || !recordBodyElementKind(kind) {
			return 0, optional, fmt.Errorf("collection elements require declared scalar or record values")
		}
		return kind + 3, optional, nil
	}
	if named, ok := valueType.(*types.Named); ok {
		if _, record := named.Underlying().(*types.Struct); record {
			return recordBodyRecordKind, optional, nil
		}
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
	case 4, 5, 6, recordBodyRecordSequenceKind:
		collection, ok := value.(*recordBodySequence)
		if !ok || collection == nil || collection.Kind != kind-3 || len(collection.Values) > RepeatedFieldLimit {
			return recordBodyScalar{}, fmt.Errorf("record field requires its declared collection element type")
		}
		scalar.Collection = collection
	case recordBodyRecordKind:
		record, ok := value.(recordBodyValue)
		if !ok || record.Count < 1 {
			return recordBodyScalar{}, fmt.Errorf("record field requires a declared record value, got %T", value)
		}
		if err := validateRecordBodyValueBudget(record); err != nil {
			return recordBodyScalar{}, err
		}
		scalar.Record = &record
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
		value = recordBodyScalarValue(pointer.Value)
		fieldType = fieldType.(*types.Pointer).Elem()
	} else if _, ok := value.(recordBodyOptionalScalar); ok {
		return recordBodyScalar{}, fmt.Errorf("required record field assignment does not accept an optional value")
	}
	if kind == recordBodyRecordKind || kind == recordBodyRecordSequenceKind {
		value, err = coerceBodyValue(value, fieldType)
		if err != nil {
			return recordBodyScalar{}, err
		}
	}
	return recordBodyScalarFromValue(value, kind)
}

func recordBodyScalarValue(value recordBodyScalar) any {
	switch value.Kind {
	case recordBodyRecordKind:
		if value.Record == nil {
			return nil
		}
		return *value.Record
	case 4, 5, 6, recordBodyRecordSequenceKind:
		return value.Collection
	case 2:
		return value.Boolean
	case 3:
		return value.Integer
	default:
		return value.Text
	}
}

func (e *integerBodyEvaluator) evaluateRecordLiteral(literal *ast.CompositeLit) (any, error) {
	if sequence, ok := e.information.Types[literal].Type.Underlying().(*types.Slice); ok {
		return e.evaluateSequenceLiteral(literal, sequence)
	}
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
	return value, validateRecordBodyValueBudget(value)
}

func (e *integerBodyEvaluator) evaluateRecordField(selector *ast.SelectorExpr) (any, error) {
	value, err := e.evaluateExpression(selector.X)
	if err != nil {
		return nil, err
	}
	if optional, ok := value.(recordBodyOptionalScalar); ok {
		if !optional.Present {
			return nil, fmt.Errorf("field read requires a present optional record")
		}
		value = recordBodyScalarValue(optional.Value)
	}
	record, ok := value.(recordBodyValue)
	if !ok {
		return nil, fmt.Errorf("field read requires a record value")
	}
	receiver := e.information.Types[selector.X].Type
	if pointer, ok := receiver.(*types.Pointer); ok {
		receiver = pointer.Elem()
	}
	structure := receiver.Underlying().(*types.Struct)
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
	if fill, ok := ctx.Value(fillCaptureKey{}).(*fillCallCapture); ok {
		return fill.record(ctx, source, activity, records, cases)
	}
	return evaluateRecordAssemblyRecorded(ctx, source, activity, records, cases, nil)
}

func evaluateRecordAssemblyRecorded(ctx context.Context, source []byte, activity string, records []RecordType,
	cases []assemblyspec.ValueCase, recorder *localCallRecorder) ([]RecordAssemblyCase, error) {
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
	e := integerBodyEvaluator{context: ctx, localCalls: recorder, information: types.Info{Types: make(map[ast.Expr]types.TypeAndValue),
		Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}}
	if _, err = new(types.Config).Check(file.Name.Name, fset, []*ast.File{file}, &e.information); err != nil {
		return nil, err
	}
	if err := ValidateRecordUseBudget(&e.information, records, false); err != nil {
		return nil, err
	}
	e.functions = pureEvaluatorFunctions(file, e.information)
	var inputs []types.Object
	for _, parameter := range function.Type.Params.List {
		for _, name := range parameter.Names {
			inputs = append(inputs, e.information.Defs[name])
		}
	}
	output := e.information.Types[function.Type.Results.List[0].Type].Type
	if output == types.Typ[types.Int64] || output == types.Typ[types.Bool] || output == types.Typ[types.String] {
		return evaluateScalarCriteriaCases(ctx, source, activity, records, cases, recorder)
	}
	named, ok := output.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("record output contract is missing")
	}
	record := recordTypeByName(records, named.Obj().Name())
	if record == nil {
		return nil, fmt.Errorf("record output contract is missing")
	}
	results := make([]RecordAssemblyCase, 0, len(cases))
	for caseIndex, c := range cases {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		e.caseIndex = caseIndex
		result, err := e.evaluateRecordCase(function, inputs, output, *record, records, c)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		if recorder != nil {
			recorder.observation.CasesCompleted++
		}
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
	e.callCount = nil
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
	if e.localCalls != nil {
		e.localCalls.observation.CasesStarted++
	}
	value, returned, err := e.evaluateBlock(function.Body)
	if err != nil || !returned {
		return RecordAssemblyCase{}, fmt.Errorf("record case did not return: %v", err)
	}
	actual, ok := value.(recordBodyValue)
	if !ok {
		return RecordAssemblyCase{}, fmt.Errorf("record case returned %T", value)
	}
	if err := validateNestedRecordOutput(actual, record, records); err != nil {
		return RecordAssemblyCase{}, err
	}
	wanted := expected.(recordBodyValue)
	result := RecordAssemblyCase{Inputs: json.RawMessage(c.Inputs), Expected: json.RawMessage(c.Expected), Passed: equalRecordBodyValue(actual, wanted)}
	object := make(map[string]any, len(record.Fields))
	for i, field := range record.Fields {
		actualScalar, expectedScalar := actual.Values[i], wanted.Values[i]
		actualValue, expectedValue := "ABSENT", "ABSENT"
		if actualScalar.Present {
			actualValue = recordScalarDisplay(actualScalar)
		}
		if expectedScalar.Present {
			expectedValue = recordScalarDisplay(expectedScalar)
		}
		typeID := ""
		if field.TypeID != "urn:gooo:type:string" {
			typeID = field.TypeID
		}
		fieldResult := RecordAssemblyField{ID: field.ID, Name: field.Name, TypeID: typeID, Expected: expectedValue,
			Actual: actualValue, Passed: equalRecordBodyScalar(actualScalar, expectedScalar)}
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
	result.Actual, err = json.Marshal(object)
	return result, err
}

func decodeRecordCaseValue(raw []byte, t types.Type, records []RecordType) (any, error) {
	decoder := newRecordCaseDecoder(t, records)
	return decoder.value(raw, t)
}

func recordScalarJSONValue(value recordBodyScalar) any {
	return recordBodyScalarValue(value)
}
