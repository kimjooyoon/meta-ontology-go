package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestFillRejectionScoringSingletonAndReplay(t *testing.T) {
	source := tinyBodyFillFixture(`return __GOOO_BODY_HOLE_value__`)
	plan := tinyBodyFillPlan("Add one.", "value",
		IRBodyFillCandidate{ID: "wrong_type", Expression: "true"},
		IRBodyFillCandidate{ID: "zero_divisor", Expression: "input / (input - input)"},
		IRBodyFillCandidate{ID: "add", Expression: "input + 1"})
	plan.TestCases = []IRBodyFillTestCase{{Input: 2, Expected: 3}}
	plan.HoldoutTestCases = []IRBodyFillTestCase{{Input: 9007199254740993, Expected: 9007199254740994}}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd}
	result, err := generateWithIRBodyFillOptions(context.Background(), "fill.gooo", source, "Choose", plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatal(err)
	}
	f := result.Report.BodyFill
	if provider.calls != 0 || f.Decision.FallbackReason != "ONLY_VALID_CANDIDATE" || f.LocalModelPredictions != 0 ||
		!f.ExternalProviderCallsKnown || f.ExternalProviderCalls != 0 || len(f.CandidateScores) != 1 || len(f.RejectedCandidates) != 2 ||
		f.RejectedCandidates[0].Stage != "TYPECHECK" || f.RejectedCandidates[1].Stage != "TRAINING_EVALUATION" ||
		f.HoldoutCaseResults[0].Actual != 9007199254740994 {
		t.Fatal(f, provider.calls)
	}
	if _, err = ReplayIRBodyFill(context.Background(), "fill.gooo", source, plan, result); err != nil {
		t.Fatal(err)
	}
	f.RejectedCandidates[0].Reason += " changed"
	if _, err = ReplayIRBodyFill(context.Background(), "fill.gooo", source, plan, result); err == nil {
		t.Fatal("changed rejection replayed")
	}
	plan.Candidates = plan.Candidates[:2]
	_, err = generateWithIRBodyFillOptions(context.Background(), "fill.gooo", source, "Choose", plan, "", "", IRBodyFillOptions{}, provider)
	var all *NoValidBodyFillCandidates
	if !errors.As(err, &all) || len(all.Observation.Rejected) != 2 || all.Observation.PlanSHA256 == "" || provider.calls != 0 {
		t.Fatal(err)
	}
}

func TestFillRejectionFilteredTinyRequest(t *testing.T) {
	source := tinyBodyFillFixture(`return __GOOO_BODY_HOLE_value__`)
	plan := tinyBodyFillPlan("Add one.", "value",
		IRBodyFillCandidate{ID: "bad", Expression: "true"},
		IRBodyFillCandidate{ID: "add", Expression: "input + 1"},
		IRBodyFillCandidate{ID: "subtract", Expression: "input - 1"})
	plan.TestCases = []IRBodyFillTestCase{{Input: 2, Expected: 3}}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAdd}
	result, err := generateWithIRBodyFillOptions(context.Background(), "fill.gooo", source, "Choose", plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatal(err)
	}
	var state irBodyFillState
	if err = json.Unmarshal([]byte(provider.request.State), &state); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 1 || len(provider.request.Question.Options) != 2 || len(state.Candidates) != 2 || len(state.RejectedCandidates) != 1 ||
		result.Report.BodyFill.SelectedCandidateID != "add" {
		t.Fatal(result.Report.BodyFill)
	}
	if _, err = ReplayIRBodyFill(context.Background(), "fill.gooo", source, plan, result); err != nil {
		t.Fatal(err)
	}
}

func TestSourceFillRejectionBindingAndHoldoutBoundary(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-rejection/budget.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "budget.gooo", source, "PlanBudget")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRBodyFill(ctx, "budget.gooo", source, "PlanBudget", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Report.BodyFill.CandidateScores) != 3 || len(result.Report.BodyFill.RejectedCandidates) != 2 {
		t.Fatal(result.Report.BodyFill)
	}
	if _, err = RealizeSourceAssembly(ctx, "budget.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"bad_cap_type", "bad_boundary_value"} {
		r, selected, err := RealizeSourceFillCandidate(ctx, "budget.gooo", source, "PlanBudget", id)
		var rejected *SourceFillCandidateRejection
		if !errors.As(err, &rejected) || selected != nil || r.Rejection == nil || r.Rejection.CandidateID != id || r.ActivityID == "" ||
			r.PlanSHA256 != result.Report.BodyFill.IRPlanSHA256 || r.SelectedSourceSHA256 == "" || r.TestCasesTotal != 0 || r.HoldoutCasesTotal != 0 || len(r.ValueCaseResults) != 0 {
			t.Fatal(r, err)
		}
	}
	// An error on a separate holdout cannot silently reject or rank a candidate.
	changed := []byte(strings.Replace(string(source), "(input.limit / (input.used - input.used)) > 0", "(input.limit / (input.used - 9)) > 0", 1))
	r, _, err := RealizeSourceFillCandidate(ctx, "budget.gooo", changed, "PlanBudget", "bad_boundary_value")
	var rejected *SourceFillCandidateRejection
	if err == nil || errors.As(err, &rejected) || r.Rejection != nil {
		t.Fatal("holdout error became selection evidence", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, err = RealizeSourceFillCandidate(cancelled, "budget.gooo", source, "PlanBudget", "bad_cap_type"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestSourceIntegerRejectionAndExternalRecordReplay(t *testing.T) {
	ctx := context.Background()
	source := []byte("package f\nnamespace f\nentity Integer id \"f://integer\"\n" + `
activity Pick(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__ + __GOOO_BODY_HOLE_pad__" assembling {
 source_fill intent "Add one." { hole "value" hole "pad"
 candidate "bad" { fill "value" "true" fill "pad" "0" }
 candidate "good" { fill "value" "input+1" fill "pad" "0" } }
 case "1" -> "2"
 holdout_case "9007199254740993" -> "9007199254740994"
}`)
	r, _, err := RealizeSourceFillCandidate(ctx, "integer.gooo", source, "Pick", "bad")
	var rejected *SourceFillCandidateRejection
	if !errors.As(err, &rejected) || r.Rejection == nil || r.TestCasesTotal != 0 {
		t.Fatal(r, err)
	}
	spec, err := SourceAssembly(ctx, "integer.gooo", source, "Pick")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRBodyFill(ctx, "integer.gooo", source, "Pick", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = RealizeSourceAssembly(ctx, "integer.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	source, activity := readRecordBodyFillFixture(t)
	plan, _, err := sourceIRBodyFillPlan("record.gooo", source, "ReviewCandidate", &activity.Assembly.Spec)
	if err != nil {
		t.Fatal(err)
	}
	plan.Candidates[0].Fills["condition"] = "input.candidate_id"
	result, err = GenerateWithIRBodyFill(ctx, "record.gooo", source, "ReviewCandidate", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Report.BodyFill.RejectedCandidates) != 1 || result.Report.BodyFill.Decision.FallbackReason != "ONLY_VALID_CANDIDATE" {
		t.Fatal(result.Report.BodyFill)
	}
	if _, err = ReplayIRBodyFill(ctx, "record.gooo", source, plan, result); err != nil {
		t.Fatal(err)
	}
	result.Report.BodyFill.Decision.Mode = "invented"
	if _, err = ReplayIRBodyFill(ctx, "record.gooo", source, plan, result); err == nil {
		t.Fatal("altered singleton decision replayed")
	}
}
