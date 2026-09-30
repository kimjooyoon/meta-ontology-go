package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const irBodySearchPlanSchema = "gooo/body-codegen-ir-search-plan/v1"
const irBodySearchStateSchema = "gooo/body-codegen-ir-search-state/v1"

func TestGenerateWithIRBodySearchShowsScoresOnlyAfterCandidateEvaluation(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := make([]IRBodyFillTestCase, 10)
	for i := range training {
		training[i] = IRBodyFillTestCase{Input: int64(-i - 1), Expected: 0}
	}
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "identity", Expression: "input"},
		{ID: "zero", Expression: "0"},
		{ID: "negate", Expression: "-input"},
	}, training, nil, 2)

	var requests []searchRequestSnapshot
	choices := []string{"identity", "zero"}
	server := newIRBodySearchServer(t, func(call int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return choices[call]
	})
	defer server.Close()

	result, err := GenerateWithIRBodySearch(
		context.Background(), "ir-search-clamp.gooo", source, "ClampNegativeToZero",
		plan, server.URL+"/v1/systemone", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if receipt == nil {
		t.Fatal("search receipt is missing")
	}
	if receipt.SelectedCandidateID != "zero" || receipt.TrainingPassed != 10 ||
		receipt.TrainingTotal != 10 || receipt.TrainingAccuracyPercent == nil || *receipt.TrainingAccuracyPercent != 100 {
		t.Fatalf("the search did not continue from the weak candidate to the passing candidate: %#v", receipt)
	}
	if !strings.Contains(result.Source, "output = 0") {
		t.Fatalf("the passing search candidate was not emitted into the activity body:\n%s", result.Source)
	}
	if receipt.AttemptedCandidates != 2 || receipt.EvaluatedCandidates != 2 || len(receipt.Attempts) != 2 || receipt.Attempts[0].CandidateID != "identity" ||
		receipt.Attempts[0].TestCasesPassed != 0 || receipt.Attempts[0].TestCasesTotal != 10 ||
		!receipt.Attempts[0].ScoringCompleted || receipt.Attempts[0].AccuracyPercent == nil || *receipt.Attempts[0].AccuracyPercent != 0 ||
		receipt.Attempts[1].CandidateID != "zero" || receipt.Attempts[1].TestCasesPassed != 10 ||
		receipt.Attempts[1].TestCasesTotal != 10 || !receipt.Attempts[1].ScoringCompleted ||
		receipt.Attempts[1].AccuracyPercent == nil || *receipt.Attempts[1].AccuracyPercent != 100 {
		t.Fatalf("attempts do not retain the exact sequential training scores: %#v", receipt.Attempts)
	}
	if len(requests) != 2 {
		t.Fatalf("Laya requests = %d, want one per attempted candidate", len(requests))
	}
	first := requests[0]
	if first.State.Schema != irBodySearchStateSchema || first.State.Stage != "choose_before_candidate_evaluation" ||
		first.State.Activity != "ClampNegativeToZero" || !strings.Contains(first.State.BodyIR, "__GOOO_BODY_HOLE_floor__") ||
		first.State.Intent != plan.Intent || first.State.TrainingTestCount != 10 ||
		first.State.TrainingSuiteSHA256 != receipt.TrainingSuiteSHA256 || len(first.State.RemainingCandidates) != 3 {
		t.Fatalf("the initial prompt did not carry the bounded IR search context: %#v", first.State)
	}
	if len(first.State.PriorAttempts) != 0 || strings.Contains(first.Raw, "test_cases_passed") || strings.Contains(first.Raw, "accuracy_percent") {
		t.Fatalf("the initial prompt exposed candidate scores before evaluation: %s", first.Raw)
	}
	if got := searchCandidateIDs(t, first.State.RemainingCandidates); !reflect.DeepEqual(got, []string{"identity", "zero", "negate"}) {
		t.Fatalf("initial choices = %#v, want the complete declared candidate set", got)
	}
	if len(requests[1].State.PriorAttempts) != 1 {
		t.Fatalf("the second prompt did not include the first completed attempt: %#v", requests[1].State.PriorAttempts)
	}
	if len(requests[1].State.RemainingCandidates) != 2 {
		t.Fatalf("the second prompt did not restrict choices to untried candidates: %#v", requests[1].State.RemainingCandidates)
	}
	if got := searchCandidateIDs(t, requests[1].State.RemainingCandidates); !reflect.DeepEqual(got, []string{"zero", "negate"}) {
		t.Fatalf("retry choices = %#v, want only untried candidates", got)
	}
	var prior struct {
		CandidateID          string                 `json:"candidate_id"`
		ScoringCompleted     bool                   `json:"scoring_completed"`
		TestCasesPassed      int                    `json:"test_cases_passed"`
		TestCasesTotal       int                    `json:"test_cases_total"`
		AccuracyPercent      *float64               `json:"accuracy_percent"`
		FailedCasesTotal     int                    `json:"failed_cases_total"`
		FailedCasesTruncated bool                   `json:"failed_cases_truncated"`
		FailedCases          []IRBodyFillCaseResult `json:"failed_cases"`
		Error                string                 `json:"error"`
	}
	if err := json.Unmarshal(requests[1].State.PriorAttempts[0], &prior); err != nil {
		t.Fatalf("decode prior attempt from second prompt: %v", err)
	}
	if prior.CandidateID != "identity" || !prior.ScoringCompleted || prior.AccuracyPercent == nil || *prior.AccuracyPercent != 0 || prior.TestCasesPassed != 0 ||
		prior.TestCasesTotal != 10 || prior.FailedCasesTotal != 10 || !prior.FailedCasesTruncated ||
		len(prior.FailedCases) != 8 || prior.FailedCases[0].Input != -1 || prior.FailedCases[7].Input != -8 || prior.Error != "" {
		t.Fatalf("second prompt omitted or changed the prior training failure: %#v", prior)
	}
}

func TestGenerateWithIRBodySearchDoesNotPrevalidateUnselectedExpressions(t *testing.T) {
	source := readIRBodySearchFixture(t)
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "zero", Expression: "0"},
		{ID: "unsupported", Expression: "input / 0"},
	}, []IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)

	var requests []searchRequestSnapshot
	server := newIRBodySearchServer(t, func(_ int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return "zero"
	})
	defer server.Close()

	result, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatalf("an unselected unsupported expression rejected the plan: %v", err)
	}
	receipt := result.Report.BodySearch
	if receipt == nil || receipt.SelectedCandidateID != "zero" || receipt.StopReason != "TRAINING_SUITE_PASSED" ||
		receipt.AttemptedCandidates != 1 || receipt.EvaluatedCandidates != 1 || receipt.UntestedCandidates != 1 || len(receipt.Attempts) != 1 ||
		receipt.Attempts[0].CandidateID != "zero" {
		t.Fatalf("search evaluated or rejected the unselected expression: %#v", receipt)
	}
	if len(requests) != 1 || len(requests[0].State.RemainingCandidates) != 2 {
		t.Fatalf("the chooser did not receive the untouched candidate set: %#v", requests)
	}
}

func TestGenerateWithIRBodySearchKeepsHoldoutHiddenAndScoresItSeparately(t *testing.T) {
	source := readIRBodySearchFixture(t)
	const sentinel int64 = -918273645
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "identity", Expression: "input"},
		{ID: "zero", Expression: "0"},
	}, []IRBodyFillTestCase{{Input: 7, Expected: 7}},
		[]IRBodyFillTestCase{{Input: sentinel, Expected: 0}}, 2)

	var requests []searchRequestSnapshot
	server := newIRBodySearchServer(t, func(_ int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return "identity"
	})
	defer server.Close()

	result, err := GenerateWithIRBodySearch(
		context.Background(), "ir-search-clamp.gooo", source, "ClampNegativeToZero",
		plan, server.URL+"/v1/systemone", "",
	)
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if receipt == nil || receipt.SelectedCandidateID != "identity" || len(receipt.Attempts) != 1 {
		t.Fatalf("search did not stop at its first training-perfect candidate: %#v", receipt)
	}
	if receipt.StopReason != "TRAINING_SUITE_PASSED" || receipt.TrainingPassed != 1 ||
		receipt.TrainingTotal != 1 || receipt.TrainingAccuracyPercent == nil || *receipt.TrainingAccuracyPercent != 100 {
		t.Fatalf("training completion was not reported separately: %#v", receipt)
	}
	if receipt.HoldoutPassed != 0 || receipt.HoldoutTotal != 1 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 0 {
		t.Fatalf("the hidden holdout score was not recorded in its own dimension: %#v", receipt)
	}
	trainingDimension := searchCompletenessDimension(t, result.Report.CompletenessReceipt, "search_training_accuracy")
	holdoutDimension := searchCompletenessDimension(t, result.Report.CompletenessReceipt, "search_holdout_accuracy")
	if trainingDimension.Status != "PASS" || holdoutDimension.Status != "PROGRESS" ||
		!containsString(result.Report.CompletenessReceipt.CoreDimensions, holdoutDimension.ID) {
		t.Fatalf("training and holdout completeness were not reported as separate core dimensions: training=%#v holdout=%#v core=%#v",
			trainingDimension, holdoutDimension, result.Report.CompletenessReceipt.CoreDimensions)
	}
	if receipt.TrainingAccuracyPercent == nil || *receipt.TrainingAccuracyPercent != 100 || receipt.GlobalBestAccuracyPercent != nil {
		t.Fatalf("holdout failure changed training accuracy or claimed an unobserved global best: %#v", receipt)
	}
	if receipt.TrainingSuiteSHA256 == "" || receipt.HoldoutSuiteSHA256 == "" ||
		receipt.TrainingSuiteSHA256 == receipt.HoldoutSuiteSHA256 {
		t.Fatalf("training and holdout suite identities are missing or conflated: %#v", receipt)
	}
	if len(requests) != 1 || strings.Contains(requests[0].Raw, fmt.Sprint(sentinel)) {
		t.Fatalf("holdout sentinel leaked into the model request: %#v", requests)
	}
}

func TestIRBodySearchHoldoutChangesPlanIdentityButNotOutput(t *testing.T) {
	source := readIRBodySearchFixture(t)
	candidates := []IRBodyFillCandidate{
		{ID: "zero", Expression: "0"},
		{ID: "identity", Expression: "input"},
	}
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}}
	firstPlan := irBodySearchPlan(candidates, training,
		[]IRBodyFillTestCase{{Input: 17, Expected: 0}}, 2)
	secondPlan := irBodySearchPlan(candidates, training,
		[]IRBodyFillTestCase{{Input: 17, Expected: 17}}, 2)

	first, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", firstPlan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", secondPlan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if first.Source != second.Source || first.Report.GeneratedDigest != second.Report.GeneratedDigest ||
		first.Report.BodySearch.SelectedCandidateID != "zero" || second.Report.BodySearch.SelectedCandidateID != "zero" {
		t.Fatalf("changing only holdout expectations changed the selected output: first=%#v second=%#v", first.Report, second.Report)
	}
	if first.Report.PlanSHA256 == second.Report.PlanSHA256 ||
		first.Report.BodySearch.IRPlanSHA256 == second.Report.BodySearch.IRPlanSHA256 {
		t.Fatalf("holdout contract changes were not bound into plan identities: first=%#v second=%#v", first.Report, second.Report)
	}
}

func TestIRBodySearchRejectsOverlappingTrainingAndHoldoutInputs(t *testing.T) {
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "zero", Expression: "0"},
		{ID: "identity", Expression: "input"},
	}, []IRBodyFillTestCase{{Input: -1, Expected: 0}},
		[]IRBodyFillTestCase{{Input: -1, Expected: 0}}, 2)

	if err := validateIRBodySearchPlan(plan); err == nil {
		t.Fatal("search plan accepted the same input in training and holdout despite matching expected values")
	}
}

func TestGenerateWithIRBodySearchWithoutProviderIsDeterministicAndHonorsAttemptBudget(t *testing.T) {
	source := readIRBodySearchFixture(t)
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "identity", Expression: "input"},
		{ID: "zero", Expression: "0"},
	}, []IRBodyFillTestCase{
		{Input: -1, Expected: 0},
		{Input: 1, Expected: 1},
	}, nil, 1)

	first, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	firstReceipt, secondReceipt := first.Report.BodySearch, second.Report.BodySearch
	if firstReceipt == nil || secondReceipt == nil {
		t.Fatalf("deterministic fallback omitted a search receipt: first=%#v second=%#v", firstReceipt, secondReceipt)
	}
	if first.Source != second.Source || first.Report.GeneratedDigest != second.Report.GeneratedDigest ||
		firstReceipt.SelectedCandidateID != secondReceipt.SelectedCandidateID ||
		firstReceipt.IRPlanSHA256 != secondReceipt.IRPlanSHA256 ||
		firstReceipt.TrainingSuiteSHA256 != secondReceipt.TrainingSuiteSHA256 {
		t.Fatalf("search fallback changed across replay: first=%#v second=%#v", firstReceipt, secondReceipt)
	}
	if len(firstReceipt.Attempts) != plan.MaxAttempts || len(secondReceipt.Attempts) != plan.MaxAttempts ||
		firstReceipt.AttemptedCandidates != 1 || secondReceipt.AttemptedCandidates != 1 ||
		firstReceipt.EvaluatedCandidates != 1 || secondReceipt.EvaluatedCandidates != 1 ||
		firstReceipt.UntestedCandidates != 1 || secondReceipt.UntestedCandidates != 1 ||
		firstReceipt.StopReason != "MAX_ATTEMPTS" || secondReceipt.StopReason != "MAX_ATTEMPTS" {
		t.Fatalf("attempt budget or untested candidate count was not retained: first=%#v second=%#v", firstReceipt, secondReceipt)
	}
	if firstReceipt.Attempts[0].CandidateID != secondReceipt.Attempts[0].CandidateID ||
		firstReceipt.Attempts[0].TestCasesPassed != secondReceipt.Attempts[0].TestCasesPassed ||
		firstReceipt.Attempts[0].TestCasesTotal != secondReceipt.Attempts[0].TestCasesTotal {
		t.Fatalf("fallback attempt trace changed across replay: first=%#v second=%#v", firstReceipt.Attempts, secondReceipt.Attempts)
	}
}

func TestGenerateWithIRBodySearchRejectsIllTypedChoiceAndContinues(t *testing.T) {
	source := readIRBodySearchFixture(t)
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "ill_typed", Expression: "true"},
		{ID: "zero", Expression: "0"},
		{ID: "negate", Expression: "-input"},
	}, []IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)

	var requests []searchRequestSnapshot
	choices := []string{"ill_typed", "zero"}
	server := newIRBodySearchServer(t, func(call int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return choices[call]
	})
	defer server.Close()

	result, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if receipt == nil || receipt.SelectedCandidateID != "zero" || len(receipt.Attempts) != 2 ||
		receipt.AttemptedCandidates != 2 || receipt.EvaluatedCandidates != 1 {
		t.Fatalf("search did not continue after rejecting the ill-typed choice: %#v", receipt)
	}
	if receipt.Attempts[0].CandidateID != "ill_typed" || receipt.Attempts[0].TypecheckPassed ||
		receipt.Attempts[0].ScoringCompleted || receipt.Attempts[0].AccuracyPercent != nil || receipt.Attempts[0].Error == "" ||
		receipt.Attempts[1].CandidateID != "zero" || !receipt.Attempts[1].TypecheckPassed {
		t.Fatalf("selected-candidate rejection was not retained before the passing attempt: %#v", receipt.Attempts)
	}
	if len(requests) != 2 || len(requests[1].State.PriorAttempts) != 1 || len(requests[1].State.RemainingCandidates) != 2 {
		t.Fatalf("the retry prompt did not include the rejected attempt: %#v", requests)
	}
	var prior struct {
		CandidateID      string          `json:"candidate_id"`
		TypecheckPassed  bool            `json:"typecheck_passed"`
		ScoringCompleted bool            `json:"scoring_completed"`
		AccuracyPercent  json.RawMessage `json:"accuracy_percent"`
		Error            string          `json:"error"`
	}
	if err := json.Unmarshal(requests[1].State.PriorAttempts[0], &prior); err != nil {
		t.Fatal(err)
	}
	if prior.CandidateID != "ill_typed" || prior.TypecheckPassed || prior.ScoringCompleted ||
		string(prior.AccuracyPercent) != "null" || prior.Error == "" {
		t.Fatalf("retry prompt omitted the typecheck failure: %#v", prior)
	}
}

func TestGenerateWithIRBodySearchAllInvalidCandidatesReturnsFailureTrace(t *testing.T) {
	source := readIRBodySearchFixture(t)
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "bool_result", Expression: "true"},
		{ID: "mixed_add", Expression: "input + true"},
		{ID: "bool_comparison", Expression: "input > 0"},
	}, []IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 3)

	var requests []searchRequestSnapshot
	choices := []string{"bool_result", "mixed_add"}
	server := newIRBodySearchServer(t, func(call int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return choices[call]
	})
	defer server.Close()

	result, err := GenerateWithIRBodySearch(context.Background(), "ir-search-clamp.gooo", source,
		"ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
	var searchErr *IRBodySearchError
	if err == nil || !errors.As(err, &searchErr) || searchErr.Receipt == nil {
		t.Fatalf("all-invalid search did not return its typed failure trace: result=%#v err=%v", result, err)
	}
	receipt := searchErr.Receipt
	if receipt.StopReason != "NO_VALID_CANDIDATE" || len(receipt.Attempts) != 3 ||
		receipt.AttemptedCandidates != 3 || receipt.EvaluatedCandidates != 0 || receipt.UntestedCandidates != 0 ||
		receipt.TrainingAccuracyPercent != nil {
		t.Fatalf("all-invalid terminal trace is incomplete: %#v", receipt)
	}
	for _, attempt := range receipt.Attempts {
		if attempt.TypecheckPassed || attempt.ScoringCompleted || attempt.AccuracyPercent != nil || attempt.Error == "" {
			t.Fatalf("invalid candidate failure was not retained: %#v", attempt)
		}
	}
	if receipt.Attempts[2].CandidateID != "bool_comparison" || receipt.Attempts[2].Decision != nil {
		t.Fatalf("the sole remaining invalid candidate was not evaluated without another provider call: %#v", receipt.Attempts[2])
	}
	if len(requests) != 2 || len(requests[1].State.PriorAttempts) != 1 {
		t.Fatalf("second invalid-choice prompt did not preserve the first failure: %#v", requests)
	}
}

type searchRequestSnapshot struct {
	Raw   string
	State struct {
		Schema              string            `json:"schema"`
		Stage               string            `json:"stage"`
		Activity            string            `json:"activity"`
		BodyIR              string            `json:"body_ir"`
		Intent              string            `json:"intent"`
		TrainingTestCount   int               `json:"training_test_count"`
		TrainingSuiteSHA256 string            `json:"training_suite_sha256"`
		RemainingCandidates []json.RawMessage `json:"remaining_candidates"`
		PriorAttempts       []json.RawMessage `json:"prior_attempts"`
	}
}

func irBodySearchPlan(candidates []IRBodyFillCandidate, training, holdout []IRBodyFillTestCase, maxAttempts int) IRBodySearchPlan {
	return IRBodySearchPlan{
		Schema: irBodySearchPlanSchema, Intent: "Clamp negative integers to zero and preserve nonnegative integers.",
		HoleID: "floor", Candidates: candidates, TestCases: training,
		HoldoutTestCases: holdout, MaxAttempts: maxAttempts,
	}
}

func searchCandidateIDs(t *testing.T, candidates []json.RawMessage) []string {
	t.Helper()
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		var value struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(candidate, &value); err != nil {
			t.Fatalf("decode remaining search candidate: %v", err)
		}
		ids = append(ids, value.ID)
	}
	return ids
}

func searchCompletenessDimension(t *testing.T, receipt *CompletenessReceipt, id string) CompletenessDimension {
	t.Helper()
	if receipt == nil {
		t.Fatal("completeness receipt is missing")
	}
	for _, dimension := range receipt.Dimensions {
		if dimension.ID == id {
			return dimension
		}
	}
	t.Fatalf("completeness dimension %q is missing: %#v", id, receipt.Dimensions)
	return CompletenessDimension{}
}

func readIRBodySearchFixture(t *testing.T) []byte {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "examples", "body-codegen", "ir-fill-clamp.gooo.fixture")
	source, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func newIRBodySearchServer(t *testing.T, choose func(call int, snapshot searchRequestSnapshot) string) *httptest.Server {
	t.Helper()
	call := 0
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" {
			http.NotFound(writer, request)
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode search request envelope: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var snapshot searchRequestSnapshot
		snapshot.Raw = payload.State["request"]
		if err := json.Unmarshal([]byte(snapshot.Raw), &snapshot.State); err != nil {
			t.Errorf("decode IR search prompt state: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		choice := choose(call, snapshot)
		call++
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"answers": map[string]any{"body_ir_search": map[string]any{"choice": choice}},
		})
	}))
}
