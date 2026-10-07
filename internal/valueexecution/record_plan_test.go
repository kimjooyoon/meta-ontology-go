package valueexecution

import (
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func recordExample(t *testing.T) ([]byte, RecordFields) {
	t.Helper()
	source, err := os.ReadFile("../../examples/language-record-binding/boolean.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/language-record-binding/boolean-input.json")
	if err != nil {
		t.Fatal(err)
	}
	fields, err := DecodeRecordInput(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, fields
}

func TestRecordPlanExecutesSourceBoundFanout(t *testing.T) {
	source, fields := recordExample(t)
	plan, err := CompileRecordPlan("main.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	first, err := plan.Execute(map[string]RecordFields{"Capture": fields})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := plan.Execute(map[string]RecordFields{"Capture": fields})
	if err != nil || !reflect.DeepEqual(first, replay) {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	if first.ApplyCalls != 3 || first.Deliveries != 2 || len(first.Results) != 3 || first.Scope != RecordTransportScope {
		t.Fatalf("execution=%+v", first)
	}
	root := first.Results["Capture"]
	for name, result := range first.Results {
		if !maps.Equal(result.Fields, fields) || result.RootInputDigest != root.RootInputDigest ||
			result.RootActivity != "Capture" || result.SourceDigest != plan.SourceDigest ||
			result.SemanticFingerprint != plan.SemanticFingerprint || result.InputOrigin != "CALLER_SUPPLIED_DATA" ||
			result.Authority != (OperationAuthority{}) || !validDigest(result.ResultDigest) || result.Fields["Complete"] != false {
			t.Fatalf("result %s=%+v", name, result)
		}
		if name != "Capture" && result.ParentResultDigest != root.ResultDigest {
			t.Fatalf("unbound predecessor for %s", name)
		}
	}
	first.Results["Review"].Fields["State"] = "CLOSED"
	fields["State"] = "REFUTED"
	if first.Results["Report"].Fields["State"] != "UNKNOWN" || root.Fields["State"] != "UNKNOWN" {
		t.Fatal("detached data aliases another result or input")
	}
	t.Log("record transport activities=3/3 deliveries=2/2 replay=1/1 unknown_preserved=3/3 authority_grants=0")
}

func TestRecordPlanRejectsInputBeforeAnyApply(t *testing.T) {
	source, fields := recordExample(t)
	plan, err := CompileRecordPlan("main.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	missing, extra := maps.Clone(fields), maps.Clone(fields)
	delete(missing, "State")
	extra["Invented"] = "value"
	cases := []map[string]RecordFields{
		{}, {"Capture": nil}, {"Capture": missing}, {"Capture": extra},
		{"Missing": fields}, {"Capture": fields, "Review": fields},
	}
	for index, inputs := range cases {
		execution, err := plan.Execute(inputs)
		if err == nil || execution.ApplyCalls != 0 || execution.Deliveries != 0 || len(execution.Results) != 0 {
			t.Fatalf("case=%d execution=%+v err=%v", index, execution, err)
		}
	}
}

func TestRecordPlanRejectsUnsupportedSource(t *testing.T) {
	source, _ := recordExample(t)
	text := string(source)
	cases := []string{
		strings.Replace(text, "record.forward:v1", "record.approve:v1", 1),
		strings.Replace(text, "type string required one", "type string optional many", 1),
		strings.Replace(text, "type string required one", "type string required many", 1),
		text + "\nbind Capture.result -> Review.input\n",
		text + "\nbind Review.result -> Capture.input\n",
	}
	for index, source := range cases {
		if _, err := CompileRecordPlan("mutant.gooo", []byte(source)); err == nil {
			t.Fatalf("source mutation %d compiled", index)
		}
	}
}

func TestRecordResultIsOpaqueAndSourceBound(t *testing.T) {
	source, fields := recordExample(t)
	plan, err := CompileRecordPlan("main.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	program := plan.programs["Capture"]
	result := issueProducedRecord(program.authority, fields, "Capture", digestValue(fields), "")
	evidence := result.Evidence()
	evidence.Fields["State"] = "CLOSED"
	if !result.Valid() || result.Evidence().Fields["State"] != "UNKNOWN" {
		t.Fatal("detached evidence changed private result")
	}
	if err := json.Unmarshal([]byte("{}"), &result); err == nil || result.Valid() {
		t.Fatal("JSON manufactured a result handle")
	}
	plan.SourceDigest = digestBytes([]byte("changed"))
	if execution, err := plan.Execute(map[string]RecordFields{"Capture": fields}); err == nil || execution.ApplyCalls != 0 {
		t.Fatal("mutated public plan authority executed")
	}
	if execution, err := (RecordPlan{}).Execute(nil); err == nil || execution.ApplyCalls != 0 {
		t.Fatal("zero-value plan executed")
	}
}

func TestRecordInputRejectsAmbiguousOrNonScalarJSON(t *testing.T) {
	for _, raw := range []string{
		"null", "[]", "1", `{"State":null}`, `{"State":1.5}`, `{"State":9223372036854775808}`, `{"State":[]}`, `{"State":{}}`,
		`{"State":"UNKNOWN","State":"CLOSED"}`, `{"State":"UNKNOWN"} {}`,
	} {
		if _, err := DecodeRecordInput([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestRecordPlanPreservesExactIntegerFieldValues(t *testing.T) {
	const source = `package integerrecords
namespace integerrecords
entity Measurement id "records://measurement" fields {
  field count id "records://measurement/count" type integer required one
  field label id "records://measurement/label" type string required one
}
activity Capture(Measurement) -> Measurement computes "record.forward:v1"
`
	plan, err := CompileRecordPlan("integer.gooo", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	fields, err := DecodeRecordInput([]byte(`{"count":9007199254740993,"label":"exact"}`))
	if err != nil || fields["count"] != int64(9007199254740993) {
		t.Fatalf("integer input lost precision: fields=%#v err=%v", fields, err)
	}
	execution, err := plan.Execute(map[string]RecordFields{"Capture": fields})
	if err != nil || execution.Results["Capture"].Fields["count"] != int64(9007199254740993) {
		t.Fatalf("integer record was not transported exactly: execution=%+v err=%v", execution, err)
	}
	wrongType := maps.Clone(fields)
	wrongType["count"] = "9007199254740993"
	if rejected, err := plan.Execute(map[string]RecordFields{"Capture": wrongType}); err == nil || rejected.ApplyCalls != 0 {
		t.Fatalf("text was accepted for an integer field: execution=%+v err=%v", rejected, err)
	}
}

func TestRecordPlanValidatesBooleanFieldsAgainstSourceType(t *testing.T) {
	source, fields := recordExample(t)
	plan, err := CompileRecordPlan("main.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	booleanTypeID := ""
	for _, field := range plan.programs["Capture"].fields {
		if field.Name == "Complete" {
			booleanTypeID = field.TypeID
		}
	}
	if booleanTypeID != string(semantic.BuiltinBooleanTypeID) {
		t.Fatalf("boolean type id=%q", booleanTypeID)
	}
	wrongType := maps.Clone(fields)
	wrongType["Complete"] = "false"
	if result, err := plan.Execute(map[string]RecordFields{"Capture": wrongType}); err == nil || result.ApplyCalls != 0 {
		t.Fatalf("wrong type executed: result=%+v err=%v", result, err)
	}
	fields["Complete"] = true
	result, err := plan.Execute(map[string]RecordFields{"Capture": fields})
	if err != nil {
		t.Fatal(err)
	}
	if result.Results["Report"].Fields["Complete"] != true {
		t.Fatalf("Boolean value not transported: %+v", result.Results["Report"].Fields)
	}
}

func TestRecordPlanPreservesOptionalScalarPresence(t *testing.T) {
	source, err := os.ReadFile("../../examples/language-record-binding/optional.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile("../../examples/language-record-binding/optional-input.json")
	if err != nil {
		t.Fatal(err)
	}
	fields, err := DecodeRecordInput(input)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := CompileRecordPlan("optional.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]RecordFields{"Capture": fields})
	if err != nil {
		t.Fatal(err)
	}
	if execution.ApplyCalls != 2 || execution.Deliveries != 1 {
		t.Fatalf("execution=%+v", execution)
	}
	for _, name := range []string{"Capture", "Relay"} {
		result := execution.Results[name].Fields
		if _, present := result["Note"]; present || result["Label"] != "" || result["Complete"] != false || result["Count"] != int64(0) {
			t.Fatalf("optional presence changed in %s: %#v", name, result)
		}
	}

	missingRequired := maps.Clone(fields)
	delete(missingRequired, "Name")
	extra := maps.Clone(fields)
	extra["Invented"] = "value"
	wrongOptionalType := maps.Clone(fields)
	wrongOptionalType["Count"] = int32(0)
	for name, invalid := range map[string]RecordFields{
		"missing-required": missingRequired,
		"extra-field":      extra,
		"wrong-optional":   wrongOptionalType,
	} {
		rejected, err := plan.Execute(map[string]RecordFields{"Capture": invalid})
		if err == nil || rejected.ApplyCalls != 0 || len(rejected.Results) != 0 {
			t.Fatalf("%s execution=%+v err=%v", name, rejected, err)
		}
	}
}

func TestRecordPlanAllowsEmptyAllOptionalRecord(t *testing.T) {
	const source = `package optionalrecords
namespace optionalrecords
entity Profile id "records://profile" fields {
  field note id "records://profile/note" type string optional one
  field active id "records://profile/active" type boolean optional one
  field count id "records://profile/count" type integer optional one
}
activity Capture(Profile) -> Profile computes "record.forward:v1"
`
	plan, err := CompileRecordPlan("optional.gooo", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	empty, err := DecodeRecordInput([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]RecordFields{"Capture": empty})
	if err != nil || execution.ApplyCalls != 1 || len(execution.Results["Capture"].Fields) != 0 {
		t.Fatalf("execution=%+v err=%v", execution, err)
	}
}
