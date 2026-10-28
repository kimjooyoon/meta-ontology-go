package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"strings"
	"testing"
	"unsafe"
)

func nominalTestType(name string, fields []*types.Var) *types.Named {
	return types.NewNamed(types.NewTypeName(token.NoPos, nil, name, nil), types.NewStruct(fields, nil), nil)
}

func nominalTestField(name string, t types.Type) *types.Var {
	return types.NewField(token.NoPos, nil, name, t, false)
}

func nominalFanout(depth, width int) *types.Named {
	child := nominalTestType("Leaf", []*types.Var{nominalTestField("value", types.Typ[types.Int64])})
	for d := 1; d < depth; d++ {
		fields := make([]*types.Var, width)
		for i := range fields {
			fields[i] = nominalTestField(fmt.Sprintf("child%d", i), child)
		}
		child = nominalTestType(fmt.Sprintf("Level%d", d), fields)
	}
	return child
}

func TestNominalRecordZeroDefaultsShareLayoutsAndChargeOccurrences(t *testing.T) {
	value, err := zeroRecordBodyValue(nominalFanout(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	if value.Values[0].Record != value.Values[1].Record {
		t.Fatal("same nominal defaults expanded into separate child allocations")
	}
	original := value.Values[0].Record
	copy := *original
	copy.Values[0] = recordBodyScalar{Kind: 3, Present: true, Integer: 99}
	value.Values[0].Record = &copy
	if value.Values[1].Record != original || original.Values[0].Integer != 0 {
		t.Fatal("replacing a child changed its shared sibling")
	}
	if _, err := zeroRecordBodyValue(nominalFanout(8, 2)); err != nil {
		t.Fatal("rejected eighth nominal level", err)
	}
	if _, err := zeroRecordBodyValue(nominalFanout(9, 2)); err == nil {
		t.Fatal("accepted ninth nominal level")
	}
	large, err := zeroRecordBodyValue(nominalFanout(8, 32))
	if err == nil || !strings.Contains(err.Error(), "8192") {
		t.Fatal("shared defaults evaded logical occurrence bound", err)
	}
	seen := map[*recordBodyValue]bool{}
	var visit func(*recordBodyValue)
	visit = func(v *recordBodyValue) {
		if v == nil || seen[v] {
			return
		}
		seen[v] = true
		for _, field := range v.Values[:v.Count] {
			visit(field.Record)
		}
	}
	visit(&large)
	if len(seen) > 8 {
		t.Fatal("default layout DAG expanded before quota check", len(seen))
	}
}

func TestNominalRecordDecoderExact8192Boundary(t *testing.T) {
	leafFields := make([]*types.Var, 31)
	leaf := RecordType{Name: "Line", ID: "urn:bound:line"}
	item := map[string]any{}
	for i := range leafFields {
		name := fmt.Sprintf("n%d", i)
		leafFields[i] = nominalTestField(name, types.Typ[types.Int64])
		leaf.Fields = append(leaf.Fields, RecordField{Name: name, ID: "urn:bound:line/" + name, TypeID: "urn:gooo:type:integer"})
		item[name] = 0
	}
	line := nominalTestType("Line", leafFields)
	fields := []*types.Var{nominalTestField("items", types.NewSlice(line))}
	parent := RecordType{Name: "Basket", ID: "urn:bound:basket", Fields: []RecordField{{Name: "items", ID: "urn:bound:items", TypeID: leaf.ID, Cardinality: "many"}}}
	payload := map[string]any{"items": make([]any, 255)}
	for i := range payload["items"].([]any) {
		payload["items"].([]any)[i] = item
	}
	for i := range 30 {
		name := fmt.Sprintf("v%d", i)
		fields = append(fields, nominalTestField(name, types.Typ[types.Int64]))
		parent.Fields = append(parent.Fields, RecordField{Name: name, ID: "urn:bound:" + name, TypeID: "urn:gooo:type:integer"})
		payload[name] = 0
	}
	fields = append(fields, nominalTestField("extra", types.NewPointer(types.Typ[types.Int64])))
	parent.Fields = append(parent.Fields, RecordField{Name: "extra", ID: "urn:bound:extra", TypeID: "urn:gooo:type:integer", Presence: "optional"})
	typeOf := nominalTestType("Basket", fields)
	raw, _ := json.Marshal(payload)
	// 1 root + 31 present fields + 255 * (1 record + 31 fields) = 8192.
	if _, err := decodeRecordCaseValue(raw, typeOf, []RecordType{parent, leaf}); err != nil {
		t.Fatal("rejected exact 8192 logical slots", err)
	}
	payload["extra"] = 0
	raw, _ = json.Marshal(payload)
	if _, err := decodeRecordCaseValue(raw, typeOf, []RecordType{parent, leaf}); err == nil || !strings.Contains(err.Error(), "8192") {
		t.Fatal("accepted 8193 logical slots", err)
	}
}

func TestNominalRecordContextDeclinesWithoutPredictionOrCases(t *testing.T) {
	r, err := ExportRecordAssemblyContext(context.Background(), "nested.gooo", nominalRecordFixture(t), "Total", false)
	if err != nil || r.Context == nil || r.Context.Status != "DECLINED_TO_DETERMINISTIC" || r.Context.Reason != "NOMINAL_RECORD_LAYOUT_FEATURE_UNAVAILABLE" || r.CandidateTests != 0 || r.ModelPredictions != 0 {
		t.Fatal("nominal layout invented a scalar model representation", r.Context, err)
	}
}

func TestNominalRecordDecoderPreservesLegacyFlatCollections(t *testing.T) {
	record := RecordType{Name: "Flat", ID: "urn:flat"}
	fields := make([]*types.Var, 32)
	payload := map[string]any{}
	for i := range fields {
		name := fmt.Sprintf("values%d", i)
		fields[i] = nominalTestField(name, types.NewSlice(types.Typ[types.Int64]))
		record.Fields = append(record.Fields, RecordField{Name: name, ID: "urn:flat/" + name, TypeID: "urn:gooo:type:integer", Cardinality: "many"})
		payload[name] = make([]int64, 256)
	}
	raw, _ := json.Marshal(payload)
	if _, err := decodeRecordCaseValue(raw, nominalTestType("Flat", fields), []RecordType{record}); err != nil {
		t.Fatal("legacy 32 x 256 scalar layout changed", err)
	}
}

func TestNominalRecordBudgetCountsEmptyOptionalObject(t *testing.T) {
	typeOf := nominalTestType("Empty", []*types.Var{nominalTestField("value", types.NewPointer(types.Typ[types.Int64]))})
	value, err := zeroRecordBodyValue(typeOf)
	if err != nil {
		t.Fatal(err)
	}
	for _, prior := range []int{8191, 8192} {
		nodes := prior
		err := chargeRecordBodyValue(value, &nodes, 1)
		if (err == nil) != (prior == 8191) {
			t.Fatalf("empty object after %d nodes: %v", prior, err)
		}
	}
}

func TestNominalRecordDecoderRejectsScalarImpersonationAndOptionalCycle(t *testing.T) {
	field := RecordField{ID: "urn:owner:child", TypeID: "urn:declared:child"}
	declared := []RecordType{{Name: "Child", ID: field.TypeID}}
	if _, err := recordCaseFieldType(field, types.Typ[types.String], declared); err == nil {
		t.Fatal("nominal field silently became a scalar")
	}
	loop := types.NewNamed(types.NewTypeName(token.NoPos, nil, "Loop", nil), nil, nil)
	loop.SetUnderlying(types.NewStruct([]*types.Var{nominalTestField("next", types.NewPointer(loop))}, nil))
	if _, err := zeroRecordBodyValue(loop); err == nil {
		t.Fatal("optional reference cycle reached default construction")
	}
}

func TestNominalRecordReferenceFitsExistingCellLayout(t *testing.T) {
	type previousScalarCell struct {
		Text       string
		Boolean    bool
		Integer    int64
		Kind       uint8
		Present    bool
		Collection *recordBodySequence
	}
	if unsafe.Sizeof(recordBodyScalar{}) > unsafe.Sizeof(previousScalarCell{}) {
		t.Fatalf("nominal reference enlarged value cell: %d > %d", unsafe.Sizeof(recordBodyScalar{}), unsafe.Sizeof(previousScalarCell{}))
	}
}
