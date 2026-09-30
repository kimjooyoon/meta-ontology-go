package bodycodegen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestGenerateWithIRBodyFillLetsLayaChooseActualBodyAndScoresTests(t *testing.T) {
	fixture, plan := readIRBodyFillInputs(t)
	var observed irBodyFillState
	var observedFields map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/systemone" || request.Method != http.MethodPost {
			http.NotFound(writer, request)
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Laya request: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observed); err != nil {
			t.Errorf("decode Gooo IR state: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observedFields); err != nil {
			t.Errorf("decode Gooo IR state fields: %v", err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model":   "ir-fill-model",
			"routing": map[string]any{"model": "ir-fill-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice":        "zero",
				"probabilities": map[string]float64{"zero": 0.9, "identity": 0.07, "negate": 0.03},
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithIRBodyFill(
		context.Background(), "ir-fill-clamp.gooo", fixture, "ClampNegativeToZero",
		plan, server.URL+"/v1/systemone", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Source, "output = 0") || !strings.Contains(result.Source, "return output") ||
		strings.Contains(result.Source, "__GOOO_BODY_HOLE_floor__") {
		t.Fatalf("Laya choice was not materialized in the emitted body:\n%s", result.Source)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.Decision.Mode != "laya" || receipt.SelectedCandidateID != "zero" ||
		receipt.SelectedExpression != "0" {
		t.Fatalf("IR body-fill decision was not retained: %#v", receipt)
	}
	if receipt.TestCasesPassed != 9 || receipt.TestCasesTotal != 9 || receipt.FunctionalAccuracyPct != 100 {
		t.Fatalf("functional accuracy is not the exact declared-suite score: %#v", receipt)
	}
	if receipt.BestCandidateID != "zero" || receipt.BestAccuracyPercent != 100 || receipt.SelectionRegretPP != 0 {
		t.Fatalf("candidate search ceiling or selection regret was misreported: %#v", receipt)
	}
	if receipt.ProposedCandidateID != "zero" || receipt.ProposedAccuracyPct != 100 ||
		receipt.SelectionAdjustment != "proposal_retained" {
		t.Fatalf("best Laya proposal was not retained: %#v", receipt)
	}
	if len(receipt.SelectedCaseResults) != 9 {
		t.Fatalf("selected test outcomes were not retained: %#v", receipt.SelectedCaseResults)
	}
	for _, score := range receipt.CandidateScores {
		switch score.ID {
		case "zero":
			if score.TestCasesPassed != 9 || score.AccuracyPercent != 100 {
				t.Fatalf("zero candidate score = %#v", score)
			}
		case "identity":
			if score.TestCasesPassed != 5 || score.AccuracyPercent != 500.0/9.0 {
				t.Fatalf("identity candidate score = %#v", score)
			}
		case "negate":
			if score.TestCasesPassed != 5 || score.AccuracyPercent != 500.0/9.0 {
				t.Fatalf("negate candidate score = %#v", score)
			}
		default:
			t.Fatalf("unexpected candidate score: %#v", score)
		}
	}
	if observed.Schema != bodyFillStateSchema || observed.Stage != "ir_ready_before_body_emission" ||
		observed.HoleID != "floor" || !strings.Contains(observed.BodyIR, "__GOOO_BODY_HOLE_floor__") ||
		observed.TestCaseCount != 9 || observed.TestSuiteSHA256 != receipt.TestSuiteSHA256 || len(observed.Candidates) != 3 {
		t.Fatalf("Laya did not receive the complete Gooo IR plan and test context: %#v", observed)
	}
	if _, sentEveryTestInput := observedFields["test_cases"]; sentEveryTestInput {
		t.Fatalf("Laya request included verbose test inputs instead of the score-and-digest summary: %#v", observedFields)
	}
	if receipt.Timing.ExecutionModel != "synchronous_sequential_no_background_codegen_goroutines" ||
		receipt.Timing.DecisionStage != "after_typed_ir_plan_and_candidate_test_scores_before_final_emission" ||
		receipt.Timing.TotalMS < receipt.Timing.LayaDecisionMS {
		t.Fatalf("Laya call order/timing was not recorded: %#v", receipt.Timing)
	}
	if receipt.IRPlanSHA256 == "" || receipt.TestSuiteSHA256 == "" || !result.Report.TypecheckPassed ||
		!result.Report.DeterministicReplay {
		t.Fatalf("body-fill proof fields are incomplete: report=%#v receipt=%#v", result.Report, receipt)
	}
}

func TestGenerateWithIRBodyFillCorrectsImperfectLayaChoiceFromTestScores(t *testing.T) {
	fixture, plan := readIRBodyFillInputs(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"answers":{"body_ir_fill":{"choice":"identity"}}}`))
	}))
	defer server.Close()

	result, err := GenerateWithIRBodyFill(
		context.Background(), "ir-fill-clamp.gooo", fixture, "ClampNegativeToZero",
		plan, server.URL+"/v1/systemone", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.ProposedCandidateID != "identity" || receipt.ProposedAccuracyPct != 500.0/9.0 ||
		receipt.SelectedCandidateID != "zero" || receipt.FunctionalAccuracyPct != 100 ||
		receipt.TestCasesPassed != 9 || receipt.TestCasesTotal != 9 ||
		receipt.Evaluator != "gooo/bodycodegen-int64-ast-interpreter/v1" || receipt.BestCandidateID != "zero" ||
		receipt.BestAccuracyPercent != 100 || receipt.SelectionRegretPP != 100-500.0/9.0 ||
		receipt.SelectionAdjustment != "replaced_with_best_scoring_candidate" || receipt.Decision.Selected != "identity" {
		t.Fatalf("imperfect model outcome was hidden or mis-scored: %#v", receipt)
	}
	if !strings.Contains(result.Source, "output = 0") || !strings.Contains(result.Source, "return output") ||
		strings.Contains(result.Source, "__GOOO_BODY_HOLE_floor__") {
		t.Fatalf("best test-scoring candidate was not emitted as actual body content:\n%s", result.Source)
	}
}

func TestGenerateWithIRBodyFillWaitsForLayaBeforeFinalEmission(t *testing.T) {
	fixture, plan := readIRBodyFillInputs(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/systemone" {
			http.NotFound(writer, request)
			return
		}
		close(entered)
		<-release
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"answers":{"body_ir_fill":{"choice":"zero"}}}`))
	}))
	defer server.Close()
	type outcome struct {
		result Result
		err    error
	}
	done := make(chan outcome, 1)
	started := time.Now()
	go func() {
		result, err := GenerateWithIRBodyFill(
			context.Background(), "ir-fill-clamp.gooo", fixture, "ClampNegativeToZero",
			plan, server.URL+"/v1/systemone", "",
		)
		done <- outcome{result: result, err: err}
	}()
	<-entered
	select {
	case completed := <-done:
		t.Fatalf("code generation returned before Laya answered: %#v", completed)
	case <-time.After(60 * time.Millisecond):
	}
	close(release)
	released = true
	completed := <-done
	if completed.err != nil {
		t.Fatal(completed.err)
	}
	if completed.result.Report.BodyFill == nil || completed.result.Report.BodyFill.Decision.Mode != "laya" {
		t.Fatalf("completed generation did not consume the synchronous Laya answer: %#v", completed.result.Report.BodyFill)
	}
	if completed.result.Report.BodyFill.Timing.LayaDecisionMS < 50 || time.Since(started) < 50*time.Millisecond {
		t.Fatalf(
			"delayed model response was not included in the critical path: %#v elapsed=%s",
			completed.result.Report.BodyFill.Timing, time.Since(started),
		)
	}
}

func TestGenerateWithIRBodyFillTimesOutWithoutDeadlockAndUsesFallback(t *testing.T) {
	fixture, plan := readIRBodyFillInputs(t)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	}))
	defer server.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	started := time.Now()
	result, err := GenerateWithIRBodyFill(
		ctx, "ir-fill-clamp.gooo", fixture, "ClampNegativeToZero",
		plan, server.URL+"/v1/systemone", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(started) > 750*time.Millisecond {
		t.Fatalf("bounded Laya timeout returned too slowly: %s", time.Since(started))
	}
	if result.Report.BodyFill == nil || result.Report.BodyFill.Decision.Mode != "deterministic_fallback" ||
		result.Report.BodyFill.Decision.FallbackReason != "PROVIDER_UNAVAILABLE" {
		t.Fatalf("timeout did not produce a deterministic fallback receipt: %#v", result.Report.BodyFill)
	}
	if result.Report.BodyFill.SelectedCandidateID != "zero" || result.Report.BodyFill.FunctionalAccuracyPct != 100 {
		t.Fatalf("fallback did not emit and score the declared first candidate: %#v", result.Report.BodyFill)
	}
}

func TestGenerateWithIRBodyFillWithoutLayaIsDeterministic(t *testing.T) {
	fixture, plan := readIRBodyFillInputs(t)
	plan.Candidates = []IRBodyFillCandidate{plan.Candidates[1], plan.Candidates[0], plan.Candidates[2]}
	first, err := GenerateWithIRBodyFill(
		context.Background(), "ir-fill-clamp.gooo", fixture, "ClampNegativeToZero", plan, "", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithIRBodyFill(
		context.Background(), "ir-fill-clamp.gooo", fixture, "ClampNegativeToZero", plan, "", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	if first.Source != second.Source || first.Report.GeneratedDigest != second.Report.GeneratedDigest ||
		first.Report.BodyFill.SelectedCandidateID != second.Report.BodyFill.SelectedCandidateID ||
		first.Report.BodyFill.Decision.RequestSHA256 != second.Report.BodyFill.Decision.RequestSHA256 {
		t.Fatalf(
			"disconnected body-fill changed across replay: first=%#v second=%#v",
			first.Report.BodyFill, second.Report.BodyFill,
		)
	}
	if first.Report.BodyFill.Decision.FallbackReason != decisionroute.FallbackNotConfigured ||
		first.Report.BodyFill.Decision.Selected != "identity" || first.Report.BodyFill.SelectedCandidateID != "zero" ||
		first.Report.BodyFill.FunctionalAccuracyPct != 100 ||
		first.Report.BodyFill.SelectionAdjustment != "replaced_with_best_scoring_candidate" {
		t.Fatalf("disconnected provider fallback reason = %q", first.Report.BodyFill.Decision.FallbackReason)
	}
}

func TestGenerateWithIRBodyFillRejectsOpenOrMismatchedCandidates(t *testing.T) {
	fixture, plan := readIRBodyFillInputs(t)
	tests := []struct {
		name string
		edit func(*IRBodyFillPlan, []byte) ([]byte, *IRBodyFillPlan)
	}{
		{name: "unsupported call", edit: func(plan *IRBodyFillPlan, source []byte) ([]byte, *IRBodyFillPlan) {
			plan.Candidates[0].Expression = "helper(input)"
			return source, plan
		}},
		{name: "missing hole", edit: func(plan *IRBodyFillPlan, source []byte) ([]byte, *IRBodyFillPlan) {
			return []byte(strings.ReplaceAll(string(source), "__GOOO_BODY_HOLE_floor__", "0")), plan
		}},
		{name: "wrong type", edit: func(plan *IRBodyFillPlan, source []byte) ([]byte, *IRBodyFillPlan) {
			plan.Candidates[0].Expression = "true"
			return source, plan
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			planCopy := plan
			planCopy.Candidates = append([]IRBodyFillCandidate(nil), plan.Candidates...)
			source, editedPlan := test.edit(&planCopy, fixture)
			result, err := GenerateWithIRBodyFill(
				context.Background(), "ir-fill-clamp.gooo", source,
				"ClampNegativeToZero", *editedPlan, "", "",
			)
			if err == nil || result.Source != "" {
				t.Fatalf("invalid body-fill input emitted source: %#v err=%v", result, err)
			}
		})
	}
}

func readIRBodyFillInputs(t *testing.T) ([]byte, IRBodyFillPlan) {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "examples", "body-codegen", "ir-fill-clamp.gooo.fixture")
	planPath := filepath.Join("..", "..", "examples", "body-codegen", "ir-fill-clamp-plan.json")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	planBytes, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	var plan IRBodyFillPlan
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		t.Fatal(err)
	}
	return fixture, plan
}

func TestEvaluateIntegerCasesMatchesGeneratedGoProfile(t *testing.T) {
	const source = `package example
//gooo:generated:start id="example://activity/clamp" kind="activity"
func Clamp(input int64) int64 {
	if input < 0 {
		return 0
	}
	return input
}
//gooo:generated:end id="example://activity/clamp" kind="activity"
`
	cases := []IRBodyFillTestCase{{Input: -4, Expected: 0}, {Input: 0, Expected: 0}, {Input: 4, Expected: 4}}
	results, passed, err := evaluateIntegerCases([]byte(source), "Clamp", cases)
	if err != nil {
		t.Fatal(err)
	}
	if passed != 3 || len(results) != len(cases) {
		t.Fatalf("integer evaluator returned results=%#v passed=%d", results, passed)
	}
}

func TestIRBodyFillPlanRequiresAClosedSchema(t *testing.T) {
	_, plan := readIRBodyFillInputs(t)
	plan.Schema = "future/free-form"
	if err := validateIRBodyFillPlan(plan); err == nil {
		t.Fatalf("open-ended schema was accepted: %v", err)
	}
	_, plan = readIRBodyFillInputs(t)
	plan.Intent = strings.Repeat("한", 2000)
	if err := validateIRBodyFillPlan(plan); err != nil {
		t.Fatalf("2000-character Korean intent was rejected: %v", err)
	}
	plan.Intent += "한"
	if err := validateIRBodyFillPlan(plan); err == nil {
		t.Fatalf("intent longer than 2000 characters was accepted: %v", err)
	}
}
