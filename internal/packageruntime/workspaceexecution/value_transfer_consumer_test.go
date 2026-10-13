package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestGoooConsumesSavedValuePathFacts6781(t *testing.T) {
	path := os.Getenv("GOOO_TRANSFER_SAVED_INPUT")
	if path == "" {
		path = "../../../docs/observations/value-transfers/transfer6373.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Report OriginUseReport `json:"report"`
	}
	if err := json.Unmarshal(raw, &saved); err != nil || saved.Report.ValueFlow == nil {
		t.Fatal(err)
	}
	facts := ProgramOriginFacts(*saved.Report.ValueFlow)
	if len(facts) != 4 {
		t.Fatal("saved origin locations", len(facts))
	}
	const source = `package explanation
namespace explain6781
entity Text id "urn:gooo:type:string"
entity Origin id "urn:explain6781:origin" fields {
 field root_activity_id id "urn:explain6781:root-activity" type string required one
 field root_id id "urn:explain6781:root" type string required one
 field source_activity_id id "urn:explain6781:activity" type string required one
 field origin_kind id "urn:explain6781:kind" type string required one
 field candidate_id id "urn:explain6781:candidate" type string required one
 field hole_id id "urn:explain6781:hole" type string required one
 field call_returns id "urn:explain6781:calls" type integer required one
 field bind_steps id "urn:explain6781:binds" type integer required one
 field source_view id "urn:explain6781:view" type string required one
 field filename id "urn:explain6781:file" type string required one
 field source_sha id "urn:explain6781:sha" type string required one
 field source_start id "urn:explain6781:start" type integer required one
 field source_end id "urn:explain6781:end" type integer required one
}
activity Explain(Origin) -> Text id "urn:explain6781:explain" computes "if input.origin_kind == \"candidate\" { if input.bind_steps > 0 { return \"후보가 패키지 연결을 거쳐 전달됨\" }; return \"후보가 직접 전달됨\" }; if input.bind_steps > 0 { return \"연결된 활동의 외부 입력\" }; return \"현재 활동의 외부 입력\""
`
	want := []string{"후보가 패키지 연결을 거쳐 전달됨", "후보가 패키지 연결을 거쳐 전달됨", "현재 활동의 외부 입력", "연결된 활동의 외부 입력"}
	var rows []map[string]any
	for i, fact := range facts {
		rows = append(rows, map[string]any{"inputs": map[string]any{"Explain": fact}, "expected": map[string]any{"Explain": want[i]}})
	}
	cases, _ := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	suite, err := bodyexecution.DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	generation, err := bodyexecution.GenerateComposition(context.Background(), "경로설명.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := bodyexecution.ExecuteComposition(context.Background(), "경로설명.gooo", []byte(source), generation, suite, "")
	if err != nil || runtime.FinitePassed != 4 || runtime.FiniteTotal != 4 || runtime.ModelCalls != 0 {
		t.Fatal(err, runtime)
	}
	if output := os.Getenv("GOOO_TRANSFER_OUTPUT"); output != "" {
		encoded, _ := json.MarshalIndent(map[string]any{"consumer_source": source, "facts": facts, "cases": suite, "consumer_runtime": runtime, "saved_source_execution": false, "synthetic_fixtures": 0}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "consumer6781.json"), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("6781: new Gooo consumer4/4, four actual saved path/location facts, original program execution0, model0")
}

func TestIntegerTupleDifferenceAndHoldout6763(t *testing.T) {
	const source = `package difference
namespace difference6763
entity Integer id "urn:gooo:type:integer"
activity Difference(Integer,Integer) -> Integer id "urn:difference6763:diff" computes "return __GOOO_BODY_HOLE_expr__" assembling {
 source_fill intent "첫 수에서 둘째 수를 뺀다" {
  hole "expr"
  candidate "sum6763" { fill "expr" "input0 + input1" }
  candidate "difference6763" { fill "expr" "input0 - input1" }
 }
 value_case "[2,8]" -> "-6"
 holdout_value_case "[100,4]" -> "96"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "difference6763", Activity: "Difference"}, Packages: []packageruntime.PackageSpec{{Path: "difference6763", Name: "difference", Sources: []packageruntime.Source{{Filename: "차이.gooo", Content: source}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"difference6763:Difference.input0":53,"difference6763:Difference.input1":17},"expected":{"difference6763:Difference":36}},{"inputs":{"difference6763:Difference.input0":-29,"difference6763:Difference.input1":11},"expected":{"difference6763:Difference":-40}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 || result.Runtime.ModelCalls != 0 {
		t.Fatal(err, result.Runtime)
	}
	flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{CandidateID: "difference6763", Limit: 32})
	if err != nil || flow.Analysis != "MAPPED" || len(flow.Rows) != 1 || result.FillOriginIssue != "" {
		t.Fatal(err, flow, result.FillOriginIssue)
	}
	if output := os.Getenv("GOOO_TRANSFER_OUTPUT"); output != "" {
		encoded, _ := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": suite, "result": result, "flow": flow}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "difference6763.json"), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("6763: new two-Integer source value_case selects subtraction, separate holdout retained, new native inputs2/2, original candidate location present, model0")
}
