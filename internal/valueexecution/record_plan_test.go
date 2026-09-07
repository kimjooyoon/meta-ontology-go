package valueexecution

import (
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"strings"
	"testing"
)

func recordExample(t *testing.T) ([]byte, RecordFields) {
	t.Helper()
	source, err := os.ReadFile("../../examples/language-record-binding/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/language-record-binding/input.json")
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
			result.Authority != (OperationAuthority{}) || !validDigest(result.ResultDigest) {
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
		strings.Replace(text, "type string required one", "type bool required one", 1),
		strings.Replace(text, "type string required one", "type string optional one", 1),
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

func TestRecordInputRejectsAmbiguousOrNonStringJSON(t *testing.T) {
	for _, raw := range []string{
		"null", "[]", "1", `{"State":null}`, `{"State":1}`, `{"State":[]}`,
		`{"State":"UNKNOWN","State":"CLOSED"}`, `{"State":"UNKNOWN"} {}`,
	} {
		if _, err := DecodeRecordInput([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
