package bodycodegen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateWithIRBodyFillUsesSeveralGoooInputs(t *testing.T) {
	const source = `package arithmetic
namespace arithmetic
entity Integer id "arithmetic://integer"
activity Combine(Integer, Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
`
	var plan IRBodyFillPlan
	const encodedPlan = `{
  "schema":"gooo/body-codegen-ir-fill-plan/v1",
  "intent":"Add the first and second integer inputs.",
  "hole_id":"value",
  "candidates":[
    {"id":"sum","expression":"input0 + input1"},
    {"id":"difference","expression":"input0 - input1"}
  ],
  "test_cases":[
    {"inputs":[2,3],"expected":5},
    {"inputs":[-1,4],"expected":3}
  ],
  "holdout_test_cases":[{"inputs":[7,11],"expected":18}]
}`
	if err := json.Unmarshal([]byte(encodedPlan), &plan); err != nil {
		t.Fatalf("decode multi-input plan: %v", err)
	}
	encodedCases, err := json.Marshal(plan.TestCases)
	if err != nil {
		t.Fatalf("encode multi-input cases: %v", err)
	}
	var roundTrip []IRBodyFillTestCase
	if err := json.Unmarshal(encodedCases, &roundTrip); err != nil || len(roundTrip) != 2 || roundTrip[0].Input != 2 ||
		len(roundTrip[0].Inputs) != 2 || roundTrip[0].Inputs[1] != 3 || roundTrip[0].Expected != 5 {
		t.Fatalf("multi-input case JSON round trip failed: cases=%s decoded=%#v err=%v", encodedCases, roundTrip, err)
	}
	var observed irBodyFillState
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/health" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"revisions": map[string]string{"multi-input-fixture": "0123456789abcdef0123456789abcdef"}})
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode model request: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observed); err != nil {
			t.Errorf("decode typed IR state: %v", err)
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "multi-input-fixture", "routing": map[string]any{"model": "multi-input-fixture"},
			"answers": map[string]any{"body_ir_fill": map[string]any{"choice": "sum"}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithIRBodyFill(context.Background(), "multi-input.gooo", []byte(source), "Combine", plan, server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatalf("generate multi-input body: %v", err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.Decision.Mode != "laya" || receipt.SelectedCandidateID != "sum" ||
		receipt.TestCasesPassed != 2 || receipt.TestCasesTotal != 2 || receipt.FunctionalAccuracyPct != 100 ||
		receipt.HoldoutCasesPassed != 1 || receipt.HoldoutCasesTotal != 1 || receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 {
		t.Fatalf("multi-input selection and finite completeness were not reported: %#v", receipt)
	}
	if len(receipt.SelectedCaseResults) != 2 || receipt.SelectedCaseResults[0].Input != 2 ||
		len(receipt.SelectedCaseResults[0].Inputs) != 2 || receipt.SelectedCaseResults[0].Inputs[1] != 3 {
		t.Fatalf("selected-case evidence omitted positional input values: %#v", receipt.SelectedCaseResults)
	}
	scores := map[string]IRBodyFillCandidateScore{}
	for _, score := range receipt.CandidateScores {
		scores[score.ID] = score
	}
	if scores["sum"].TestCasesPassed != 2 || scores["sum"].AccuracyPercent != 100 ||
		scores["difference"].TestCasesPassed != 0 || scores["difference"].AccuracyPercent != 0 {
		t.Fatalf("candidate metrics do not distinguish the declared body choices: %#v", scores)
	}
	if !strings.Contains(result.Source, "func Combine(input0 int64, input1 int64)") || strings.Contains(result.Source, "__GOOO_BODY_HOLE_value__") {
		t.Fatalf("generated body did not preserve both source parameters:\n%s", result.Source)
	}
	if !strings.EqualFold(observed.InputType, "Integer") || len(observed.InputTypes) != 2 || observed.InputTypes[0] != "Integer" || observed.InputTypes[1] != "Integer" {
		t.Fatalf("model context omitted the typed multi-input signature: %#v", observed)
	}
	probes := receipt.BehavioralProbes
	if probes == nil || probes.Schema != "gooo/ir-body-fill-behavioral-probe/v2" || len(probes.ProbeVectors) == 0 ||
		probes.CandidatePairsDistinguished != 1 || probes.CandidatePairDistinguishabilityPercent != 100 {
		t.Fatalf("multi-input behavioral probe vectors were not recorded: %#v", probes)
	}
	for _, vector := range probes.ProbeVectors {
		if len(vector) != 2 || vector[0] == 7 && vector[1] == 11 {
			t.Fatalf("probe vector has wrong arity or reused holdout data: %v", vector)
		}
	}
}

func TestIRBodyFillRejectsInconsistentInputVectorArity(t *testing.T) {
	var plan IRBodyFillPlan
	err := json.Unmarshal([]byte(`{
      "schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"combine","hole_id":"value",
      "candidates":[{"id":"sum","expression":"input0 + input1"},{"id":"identity","expression":"input0"}],
      "test_cases":[{"inputs":[1,2],"expected":3},{"inputs":[1],"expected":1}]
    }`), &plan)
	if err != nil {
		t.Fatalf("decode plan before semantic validation: %v", err)
	}
	if err := validateIRBodyFillPlan(plan); err == nil || !strings.Contains(err.Error(), "consistent") {
		t.Fatalf("inconsistent input vector arity was accepted: %v", err)
	}
}
