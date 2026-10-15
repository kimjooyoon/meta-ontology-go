package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func fillLocalFixture() ([]byte, IRBodyFillPlan) {
	source := []byte(`package fillhistory
namespace fillhistory
entity Integer id "fillhistory://integer"
activity Step(Integer) -> Integer computes "return 12 / input"
activity Build(Integer) -> Integer computes "return Step(__GOOO_BODY_HOLE_value__)"
`)
	plan := tinyBodyFillPlan("Select a working divisor.", "value",
		IRBodyFillCandidate{ID: "fails", Expression: "input - 2"},
		IRBodyFillCandidate{ID: "works", Expression: "input"})
	plan.TestCases = []IRBodyFillTestCase{{Input: 4, Expected: 3}, {Input: 2, Expected: 6}}
	plan.HoldoutTestCases = []IRBodyFillTestCase{{Input: 3, Expected: 4}}
	return source, plan
}

func TestFillLocalCallsRetainFailedPrefixAndSelectedPasses(t *testing.T) {
	source, plan := fillLocalFixture()
	r, err := GenerateWithIRBodyFill(context.Background(), "fill.gooo", source, "Build", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(r.Report.BodyFill)
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		LocalCalls *LocalCallHistory `json:"local_calls"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.LocalCalls == nil || len(receipt.LocalCalls.Evaluations) != 5 {
		t.Fatal("fill evaluation passes were not retained", receipt.LocalCalls)
	}
	wantPhases := []string{"fill-training", "fill-training", "fill-probe", "fill-selected-training", "fill-selected-holdout"}
	for i, evaluation := range receipt.LocalCalls.Evaluations {
		if evaluation.Phase != wantPhases[i] || evaluation.ActivityID != r.Report.ActivityID ||
			evaluation.GeneratedSHA256 == "" || evaluation.SuiteSHA256 == "" {
			t.Fatal("unbound or reordered pass", i, evaluation)
		}
	}
	failed := receipt.LocalCalls.Evaluations[0]
	if failed.CandidateID != "fails" || failed.EvaluationCompleted || failed.CasesStarted != 2 || failed.CasesCompleted != 1 ||
		len(failed.Calls) != 2 || string(failed.Calls[0].Inputs) != "[2]" || string(failed.Calls[1].Inputs) != "[0]" {
		t.Fatal("failed candidate lost entered calls", failed)
	}
	last := receipt.LocalCalls.Evaluations[4]
	if last.CandidateID != "works" || len(last.Calls) != 1 || string(last.Calls[0].Inputs) != "[3]" {
		t.Fatal("selected holdout call missing", last)
	}
}

func TestFillLocalCallsReplayPreservesLegacyAbsence(t *testing.T) {
	source, plan := fillLocalFixture()
	ctx := context.Background()
	r, err := GenerateWithIRBodyFill(ctx, "fill.gooo", source, "Build", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, err := ReplayIRBodyFill(ctx, "fill.gooo", source, plan, saved); err != nil {
		t.Fatal("saved typed history did not replay", err)
	}
	for _, change := range []func(*LocalCallHistory){
		func(h *LocalCallHistory) { h.Evaluations[0].Calls[1].Inputs = json.RawMessage("[1]") },
		func(h *LocalCallHistory) { h.Evaluations[0].Phase = "fill-selected-training" },
		func(h *LocalCallHistory) { h.Evaluations[0].CasesCompleted++ },
		func(h *LocalCallHistory) { h.Evaluations = h.Evaluations[1:] },
	} {
		if err := json.Unmarshal(raw, &saved); err != nil {
			t.Fatal(err)
		}
		change(saved.Report.BodyFill.LocalCalls)
		if _, err := ReplayIRBodyFill(ctx, "fill.gooo", source, plan, saved); err == nil {
			t.Fatal("changed helper history replayed")
		}
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	saved.Report.BodyFill.LocalCalls = nil
	if _, err := ReplayIRBodyFill(ctx, "fill.gooo", source, plan, saved); err != nil || saved.Report.BodyFill.LocalCalls != nil {
		t.Fatal("legacy absence was rejected or filled in", err)
	}
}

func TestFillLocalCallsKeepAllRejectedPrefix(t *testing.T) {
	source, plan := fillLocalFixture()
	plan.Candidates[1].Expression = "0"
	r, err := GenerateWithIRBodyFill(context.Background(), "fill.gooo", source, "Build", plan, "", "")
	var failed *NoValidBodyFillCandidates
	if !errors.As(err, &failed) || r.Source != "" || failed.Observation.LocalCalls == nil ||
		len(failed.Observation.LocalCalls.Evaluations) != 2 {
		t.Fatal("failed fill lost its bounded call history", err)
	}
	for _, pass := range failed.Observation.LocalCalls.Evaluations {
		if pass.EvaluationCompleted || len(pass.Calls) == 0 || pass.Phase != "fill-training" {
			t.Fatal("failed pass gained a score or lost its entered prefix", pass)
		}
	}
}

func TestFillLocalCallsBindRecordCandidatesAndExactValues(t *testing.T) {
	source := []byte(`package recordhistory
namespace recordhistory
entity Box id "recordhistory://box" fields {
 field amount id "recordhistory://amount" type integer required one
 field note id "recordhistory://note" type string optional one
}
activity Step(Box) -> Box computes "return input"
activity Build(Box) -> Box computes "return __GOOO_BODY_HOLE_value__"
`)
	plan := IRBodyFillPlan{Schema: bodyFillRecordPlanSchema, Intent: "Preserve the record.",
		Holes: []IRBodyFillHole{{ID: "value"}}, Candidates: []IRBodyFillCandidate{
			{ID: "zero", Fills: map[string]string{"value": "Box{amount:0}"}},
			{ID: "helper", Fills: map[string]string{"value": "Step(input)"}},
		}, ValueCases: []assemblyspec.ValueCase{{Inputs: `[{"amount":9007199254740993}]`,
			Expected: `{"amount":9007199254740993}`}},
		ValueHoldoutCases: []assemblyspec.ValueCase{{Inputs: `[{"amount":5,"note":""}]`, Expected: `{"amount":5,"note":""}`}}}
	ctx := context.Background()
	r, err := GenerateWithIRBodyFill(ctx, "record.gooo", source, "Build", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	history := r.Report.BodyFill.LocalCalls
	if history == nil || history.Status != "RECORDED" || len(history.Evaluations) != 4 ||
		len(history.Evaluations[0].Calls) != 0 || len(history.Evaluations[1].Calls) != 1 ||
		string(history.Evaluations[1].Calls[0].Inputs) != `[{"amount":9007199254740993,"note":null}]` ||
		string(history.Evaluations[3].Calls[0].Inputs) != plan.ValueHoldoutCases[0].Inputs {
		t.Fatal("candidate-only helper, exact integer or optional presence lost", history)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, err := ReplayIRBodyFill(ctx, "record.gooo", source, plan, saved); err != nil {
		t.Fatal("record helper history did not replay", err)
	}
}

func TestFillLocalCallsBudgetAndBindingLeaveScoresIntact(t *testing.T) {
	source, plan := fillLocalFixture()
	completed := []byte(strings.Replace(string(source), "__GOOO_BODY_HOLE_value__", "input", 1))
	r, err := Generate("fill.gooo", completed, "Build")
	if err != nil {
		t.Fatal(err)
	}
	ctx, capture := captureFillCalls(context.Background())
	capture.bind([]byte(r.Source), r.Report)
	capture.calls = localCallRecordLimit - 1
	_, passed, err := capture.integer(localCallPhase(ctx, "fill-training", "works"), []byte(r.Source), "Build", plan.TestCases)
	if err != nil || passed != 2 || capture.history.Status != "UNKNOWN" || capture.history.Reason != "CALL_RECORD_LIMIT" ||
		len(capture.history.Evaluations[0].Calls) != 1 {
		t.Fatal("bounded capture changed evaluation or claimed completeness", capture.history, err)
	}
	ctx, unbound := captureFillCalls(context.Background())
	_, passed, err = unbound.integer(localCallPhase(ctx, "fill-training", "works"), []byte(r.Source), "Build", plan.TestCases)
	if err != nil || passed != 2 || unbound.history.Status != "UNKNOWN" ||
		unbound.history.Reason != "EVALUATION_BINDING_UNAVAILABLE" {
		t.Fatal("unbound source gained an observed identity", unbound.history, err)
	}
}

func TestFillLocalCallsRecordVectorProbes(t *testing.T) {
	source := []byte(`package vectorhistory
namespace vectorhistory
entity Integer id "vectorhistory://integer"
activity Step(Integer, Integer) -> Integer computes "return input0 + input1"
activity Build(Integer, Integer) -> Integer computes "return Step(__GOOO_BODY_HOLE_value__, input1)"
`)
	plan := tinyBodyFillPlan("Add both arguments.", "value",
		IRBodyFillCandidate{ID: "sum", Expression: "input0"}, IRBodyFillCandidate{ID: "zero", Expression: "0"})
	plan.TestCases = []IRBodyFillTestCase{{Input: 2, Inputs: []int64{2, 3}, Expected: 5}}
	plan.HoldoutTestCases = []IRBodyFillTestCase{{Input: 10, Inputs: []int64{10, 1}, Expected: 11}}
	r, err := GenerateWithIRBodyFill(context.Background(), "vector.gooo", source, "Build", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	probes := 0
	for _, pass := range r.Report.BodyFill.LocalCalls.Evaluations {
		if pass.Phase != "fill-probe" {
			continue
		}
		probes++
		if len(pass.Calls) == 0 || pass.CandidateID == "" || pass.Status != "RECORDED" {
			t.Fatal("vector probe input history missing", pass)
		}
		for _, call := range pass.Calls {
			var tuple []int64
			if err := json.Unmarshal(call.Inputs, &tuple); err != nil || len(tuple) != 2 {
				t.Fatal("vector argument order lost", string(call.Inputs), err)
			}
		}
	}
	if probes != 2 {
		t.Fatal("expected one probe pass per candidate", probes)
	}
}

func TestFillCallerLocalHistoryIncludesRejectedPrefix(t *testing.T) {
	source, _ := fillLocalFixture()
	source = []byte(strings.TrimSpace(strings.Replace(string(source), "Step(__GOOO_BODY_HOLE_value__)",
		"Step(__GOOO_BODY_HOLE_value__) + __GOOO_BODY_HOLE_bias__", 1)) + ` assembling {
 source_fill intent "Select a working divisor." {
  hole "value"
  hole "bias"
  candidate "fails" { fill "value" "input - 2" fill "bias" "0" }
  candidate "works" { fill "value" "input" fill "bias" "0" }
 }
 case "4" -> "3"
 case "2" -> "6"
 holdout_case "3" -> "4"
}
`)
	ctx := context.Background()
	failed, _, err := RealizeSourceFillCandidate(ctx, "caller.gooo", source, "Build", "fails")
	var rejection *SourceFillCandidateRejection
	if !errors.As(err, &rejection) || failed.LocalCalls == nil || len(failed.LocalCalls.Evaluations) != 1 ||
		len(failed.LocalCalls.Evaluations[0].Calls) != 2 || failed.TestCasesTotal != 0 || failed.HoldoutCasesTotal != 0 {
		t.Fatal("caller rejection lost history or claimed a completed suite", failed, err)
	}
	good, _, err := RealizeSourceFillCandidate(ctx, "caller.gooo", source, "Build", "works")
	if err != nil || good.LocalCalls == nil || len(good.LocalCalls.Evaluations) != 2 ||
		good.LocalCalls.Evaluations[1].Phase != "fill-caller-holdout" {
		t.Fatal("caller holdout was not recorded", good, err)
	}
	original := []FillCandidate{failed, good}
	raw, _ := json.MarshalIndent(original, "", "  ")
	var saved []FillCandidate
	if err := json.Unmarshal(raw, &saved); err != nil || !EqualFillCandidateHistory(saved, original) {
		t.Fatal("caller history does not survive JSON", err)
	}
	saved[0].LocalCalls.Evaluations[0].Calls[1].Inputs = json.RawMessage("[99]")
	if EqualFillCandidateHistory(saved, original) {
		t.Fatal("changed rejected caller prefix accepted")
	}
	saved[0].LocalCalls, saved[1].LocalCalls = nil, nil
	if !EqualFillCandidateHistory(saved, original) {
		t.Fatal("legacy missing caller history was upgraded or rejected")
	}
}
