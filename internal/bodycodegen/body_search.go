package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const bodySearchPlanSchema = "gooo/body-codegen-ir-search-plan/v1"

// IRBodySearchPlan bounds candidate exploration. Holdout cases are withheld
// from the chooser and measured only after the final candidate is committed.
type IRBodySearchPlan struct {
	Schema                   string                           `json:"schema"`
	Intent                   string                           `json:"intent"`
	HoleID                   string                           `json:"hole_id"`
	Candidates               []IRBodyFillCandidate            `json:"candidates"`
	CandidateGeneration      *IRBodySearchCandidateGeneration `json:"candidate_generation,omitempty"`
	TestCases                []IRBodyFillTestCase             `json:"test_cases"`
	HoldoutTestCases         []IRBodyFillTestCase             `json:"holdout_test_cases,omitempty"`
	MaxAttempts              int                              `json:"max_attempts"`
	ProviderModel            string                           `json:"provider_model,omitempty"`
	ExternalTrainingFeedback *ExternalTrainingFeedback        `json:"external_training_feedback,omitempty"`
	PromptProfile            string                           `json:"prompt_profile,omitempty"`
}

// UnmarshalJSON rejects unknown fields and case-insensitive duplicate keys
// across the full plan, including fields added to this plan in the future.
func (plan *IRBodySearchPlan) UnmarshalJSON(data []byte) error {
	type planWire IRBodySearchPlan
	var decoded planWire
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*plan = IRBodySearchPlan(decoded)
	return nil
}

type IRBodySearchAttempt struct {
	CandidateID       string                 `json:"candidate_id"`
	Expression        string                 `json:"expression"`
	SelectionMethod   string                 `json:"selection_method"`
	Decision          *decisionroute.Receipt `json:"decision,omitempty"`
	TypecheckPassed   bool                   `json:"typecheck_passed"`
	ScoringCompleted  bool                   `json:"scoring_completed"`
	TestCasesPassed   int                    `json:"test_cases_passed"`
	TestCasesTotal    int                    `json:"test_cases_total"`
	AccuracyPercent   *float64               `json:"accuracy_percent"`
	CaseResults       []IRBodyFillCaseResult `json:"case_results,omitempty"`
	Error             string                 `json:"error,omitempty"`
	DecisionLatencyMS float64                `json:"decision_latency_ms"`
	EvaluationMS      float64                `json:"evaluation_ms"`
}

// IRBodySearchReceipt records candidate evidence and provider resolver activity.
// ProviderOperations counts calls with an endpoint, including canceled calls, rather than HTTP requests.
type IRBodySearchReceipt struct {
	Schema                      string                                  `json:"schema"`
	IRPlanSHA256                string                                  `json:"ir_plan_sha256"`
	OriginalSourceDigest        string                                  `json:"original_source_digest"`
	TrainingSuiteSHA256         string                                  `json:"training_suite_sha256"`
	HoldoutSuiteSHA256          string                                  `json:"holdout_suite_sha256,omitempty"`
	SelectedCandidateID         string                                  `json:"selected_candidate_id"`
	SelectedExpression          string                                  `json:"selected_expression"`
	TrainingPassed              int                                     `json:"training_passed"`
	TrainingTotal               int                                     `json:"training_total"`
	TrainingAccuracyPercent     *float64                                `json:"training_accuracy_percent"`
	TrainingCaseResults         []IRBodyFillCaseResult                  `json:"training_case_results,omitempty"`
	HoldoutPassed               int                                     `json:"holdout_passed"`
	HoldoutTotal                int                                     `json:"holdout_total"`
	HoldoutAccuracyPercent      *float64                                `json:"holdout_accuracy_percent"`
	HoldoutCaseResults          []IRBodyFillCaseResult                  `json:"holdout_case_results,omitempty"`
	HoldoutError                string                                  `json:"holdout_error,omitempty"`
	Attempts                    []IRBodySearchAttempt                   `json:"attempts"`
	CandidateCount              int                                     `json:"candidate_count"`
	CandidateGeneration         *IRBodySearchCandidateGenerationReceipt `json:"candidate_generation,omitempty"`
	AttemptedCandidates         int                                     `json:"attempted_candidates"`
	EvaluatedCandidates         int                                     `json:"evaluated_candidates"`
	ProviderOperations          int                                     `json:"provider_operations"`
	UntestedCandidates          int                                     `json:"untested_candidates"`
	BestObservedAccuracyPercent *float64                                `json:"best_observed_accuracy_percent"`
	GlobalBestAccuracyPercent   *float64                                `json:"global_best_accuracy_percent"`
	StopReason                  string                                  `json:"stop_reason"`
	Evaluator                   string                                  `json:"evaluator"`
	ProviderBudgetMS            float64                                 `json:"provider_budget_ms"`
	ProviderBudgetUsedMS        float64                                 `json:"provider_budget_used_ms"`
	DecisionLatencyMS           float64                                 `json:"decision_latency_ms"`
	TotalMS                     float64                                 `json:"total_ms"`
	ExternalTrainingFeedback    *IRBodySearchExternalFeedbackReceipt    `json:"external_training_feedback,omitempty"`
	PromptProfile               string                                  `json:"prompt_profile,omitempty"`
}

// IRBodySearchError preserves attempted candidates even when no projection can
// be emitted. It also keeps caller cancellation distinct from provider timeout.
type IRBodySearchError struct {
	Receipt *IRBodySearchReceipt
	Cause   error
}

func (e *IRBodySearchError) Error() string { return e.Cause.Error() }
func (e *IRBodySearchError) Unwrap() error { return e.Cause }

type bodySearchFeedback struct {
	CandidateID          string                 `json:"candidate_id"`
	TypecheckPassed      bool                   `json:"typecheck_passed"`
	ScoringCompleted     bool                   `json:"scoring_completed"`
	TestCasesPassed      int                    `json:"test_cases_passed"`
	TestCasesTotal       int                    `json:"test_cases_total"`
	AccuracyPercent      *float64               `json:"accuracy_percent"`
	Error                string                 `json:"error,omitempty"`
	FailedCases          []IRBodyFillCaseResult `json:"failed_cases,omitempty"`
	FailedCasesTotal     int                    `json:"failed_cases_total"`
	FailedCasesTruncated bool                   `json:"failed_cases_truncated"`
}

// GenerateWithIRBodySearch asks for a candidate before evaluating it, and sends
// failed training observations into the next round. No candidate is repeated.
func GenerateWithIRBodySearch(ctx context.Context, filename string, source []byte, activityName string,
	plan IRBodySearchPlan, endpoint, apiKey string,
) (Result, error) {
	return generateWithIRBodySearchBudget(ctx, filename, source, activityName, plan, endpoint, apiKey, irBodyFillDecisionBudget)
}

func generateWithIRBodySearchBudget(ctx context.Context, filename string, source []byte, activityName string,
	plan IRBodySearchPlan, endpoint, apiKey string, providerBudget time.Duration,
) (Result, error) {
	started := time.Now()
	endpoint = strings.TrimSpace(endpoint)
	if err := validateIRBodySearchPlan(plan); err != nil {
		return Result{}, err
	}
	candidateGeneration, err := generateIRBodySearchCandidates(&plan)
	if err != nil {
		return Result{}, err
	}
	if feedback := plan.ExternalTrainingFeedback; feedback != nil {
		found := false
		for _, candidate := range plan.Candidates {
			if candidate.ID == feedback.CandidateID {
				found = true
				break
			}
		}
		if !found {
			return Result{}, fmt.Errorf("external training feedback candidate %q is not in the generated candidate set", feedback.CandidateID)
		}
	}
	if err := validateExternalTrainingFeedback(plan, source); err != nil {
		return Result{}, err
	}
	planBytes, _ := json.Marshal(plan)
	trainingBytes, _ := json.Marshal(plan.TestCases)
	receipt := &IRBodySearchReceipt{
		Schema: bodySearchPlanSchema, IRPlanSHA256: digest(planBytes), OriginalSourceDigest: digest(source),
		TrainingSuiteSHA256: digest(trainingBytes), TrainingTotal: len(plan.TestCases),
		HoldoutTotal: len(plan.HoldoutTestCases), CandidateCount: len(plan.Candidates),
		UntestedCandidates: len(plan.Candidates), Attempts: []IRBodySearchAttempt{},
		Evaluator: bodyFillEvaluator, ProviderBudgetMS: float64(providerBudget) / float64(time.Millisecond),
		PromptProfile: plan.PromptProfile, CandidateGeneration: candidateGeneration,
	}
	if plan.ExternalTrainingFeedback != nil {
		receipt.ExternalTrainingFeedback = externalTrainingFeedbackReceipt(plan.ExternalTrainingFeedback)
	}
	if len(plan.HoldoutTestCases) > 0 {
		cases, _ := json.Marshal(plan.HoldoutTestCases)
		receipt.HoldoutSuiteSHA256 = digest(cases)
	}
	fail := func(reason string, cause error) (Result, error) {
		receipt.StopReason, receipt.TotalMS = reason, elapsedMS(started)
		return Result{}, &IRBodySearchError{Receipt: receipt, Cause: cause}
	}
	if err := ctx.Err(); err != nil {
		return fail("CALLER_CANCELLED", err)
	}
	file, activity, activityID, body, err := prepareBodySearch(filename, source, activityName, plan.HoleID)
	if err != nil {
		return fail("INVALID_SOURCE", err)
	}
	remainingProviderBudget := providerBudget
	remaining := append([]IRBodyFillCandidate(nil), plan.Candidates...)
	bestPassed, bestBody := -1, ""
	var bestCases []IRBodyFillCaseResult
	for len(receipt.Attempts) < plan.MaxAttempts && len(remaining) > 0 {
		if err := ctx.Err(); err != nil {
			return fail("CALLER_CANCELLED", err)
		}
		attempt := IRBodySearchAttempt{TestCasesTotal: len(plan.TestCases)}
		chosen := remaining[0]
		if len(remaining) == 1 {
			attempt.SelectionMethod = "sole_remaining_candidate"
		} else {
			request, err := bodySearchRequest(activityName, activityID, body, plan, remaining, receipt)
			if err != nil {
				return fail("INVALID_REQUEST", err)
			}
			decisionStarted := time.Now()
			url := endpoint
			if remainingProviderBudget <= 0 {
				url = ""
			}
			decisionContext := ctx
			cancel := func() {}
			if url != "" {
				decisionContext, cancel = context.WithTimeout(ctx, remainingProviderBudget)
			}
			if url != "" {
				receipt.ProviderOperations++
			}
			decision, err := decisionroute.Resolve(decisionContext, request, url, apiKey)
			budgetExpired := decisionContext.Err() == context.DeadlineExceeded
			cancel()
			if url != "" {
				used := time.Since(decisionStarted)
				remainingProviderBudget -= used
				receipt.ProviderBudgetUsedMS += float64(used) / float64(time.Millisecond)
			}
			attempt.DecisionLatencyMS = elapsedMS(decisionStarted)
			receipt.DecisionLatencyMS += attempt.DecisionLatencyMS
			if err := ctx.Err(); err != nil {
				return fail("CALLER_CANCELLED", err)
			}
			if err != nil {
				return fail("INVALID_REQUEST", err)
			}
			if endpoint != "" && decision.Mode != "laya" && (url == "" || budgetExpired) {
				decision.FallbackReason = "SEARCH_PROVIDER_BUDGET_EXHAUSTED"
			}
			attempt.Decision, attempt.SelectionMethod = &decision, decision.Mode
			var found bool
			chosen, found = candidateByID(remaining, decision.Selected)
			if !found {
				return fail("INVALID_REQUEST", fmt.Errorf("search selected an undeclared or attempted candidate %q", decision.Selected))
			}
		}
		attempt.CandidateID, attempt.Expression = chosen.ID, chosen.Expression
		evaluationStarted := time.Now()
		candidateBody, results, passed, typed, err := evaluateSearchCandidate(ctx, file.Package.Name, activityName,
			activityID, body, bodyFillHoleToken(plan.HoleID), chosen.Expression, plan.TestCases)
		attempt.EvaluationMS, attempt.TypecheckPassed = elapsedMS(evaluationStarted), typed
		if err != nil {
			attempt.Error = err.Error()
		} else {
			attempt.ScoringCompleted = true
			receipt.EvaluatedCandidates++
			attempt.CaseResults, attempt.TestCasesPassed = results, passed
			accuracy := float64(passed) * 100 / float64(len(plan.TestCases))
			attempt.AccuracyPercent = &accuracy
			if passed > bestPassed {
				bestPassed, bestBody = passed, candidateBody
				bestCases = results
				receipt.SelectedCandidateID, receipt.SelectedExpression = chosen.ID, chosen.Expression
				receipt.BestObservedAccuracyPercent = &accuracy
			}
		}
		receipt.Attempts = append(receipt.Attempts, attempt)
		remaining = removeBodySearchCandidate(remaining, chosen.ID)
		receipt.AttemptedCandidates, receipt.UntestedCandidates = len(receipt.Attempts), len(remaining)
		if err := ctx.Err(); err != nil {
			return fail("CALLER_CANCELLED", err)
		}
		if err == nil && passed == len(plan.TestCases) {
			receipt.StopReason = "TRAINING_SUITE_PASSED"
			break
		}
	}
	if err := ctx.Err(); err != nil {
		return fail("CALLER_CANCELLED", err)
	}
	if bestPassed < 0 {
		return fail("NO_VALID_CANDIDATE", fmt.Errorf("body search attempted %d candidates without a valid scored body", len(receipt.Attempts)))
	}
	if receipt.StopReason == "" {
		receipt.StopReason = "MAX_ATTEMPTS"
		if len(remaining) == 0 {
			receipt.StopReason = "CANDIDATE_SET_EXHAUSTED"
		}
	}
	completedSource, err := replaceActivityProgram(source, activity.ValueProgramSpan, bestBody)
	if err != nil {
		return fail("EMISSION_FAILED", err)
	}
	result, err := GenerateWithPlanner(ctx, filename, completedSource, activityName, "", "")
	if err := ctx.Err(); err != nil {
		return fail("CALLER_CANCELLED", err)
	}
	if err != nil {
		return fail("EMISSION_FAILED", err)
	}
	finalTraining, finalPassed, err := evaluateIntegerCasesContext(ctx, []byte(result.Source), activityName, plan.TestCases)
	if err := ctx.Err(); err != nil {
		return fail("CALLER_CANCELLED", err)
	}
	if err != nil {
		return fail("EMISSION_SCORE_MISMATCH", fmt.Errorf("emitted training body could not be rescored: %w", err))
	}
	receipt.TrainingCaseResults, receipt.TrainingPassed = finalTraining, finalPassed
	trainingAccuracy := float64(finalPassed) * 100 / float64(receipt.TrainingTotal)
	receipt.TrainingAccuracyPercent = &trainingAccuracy
	if finalPassed != bestPassed || !equalIRBodyFillCaseResults(finalTraining, bestCases) {
		return fail("EMISSION_SCORE_MISMATCH", fmt.Errorf("emitted training observations differ from the selected candidate"))
	}
	// The selection is now final. Holdout results never feed a choice request.
	if len(plan.HoldoutTestCases) > 0 {
		receipt.HoldoutCaseResults, receipt.HoldoutPassed, err = evaluateIntegerCasesContext(ctx, []byte(result.Source), activityName, plan.HoldoutTestCases)
		if err != nil {
			receipt.HoldoutError = err.Error()
		} else {
			accuracy := float64(receipt.HoldoutPassed) * 100 / float64(receipt.HoldoutTotal)
			receipt.HoldoutAccuracyPercent = &accuracy
		}
	}
	if err := ctx.Err(); err != nil {
		return fail("CALLER_CANCELLED", err)
	}
	receipt.TotalMS = elapsedMS(started)
	result.Report.BodySearch = receipt
	populateCompletenessReceipt(&result.Report, "")
	return result, nil
}

func evaluateSearchCandidate(ctx context.Context, packageName, activity, activityID, body, hole, expression string,
	cases []IRBodyFillTestCase,
) (string, []IRBodyFillCaseResult, int, bool, error) {
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return "", nil, 0, false, err
	}
	if err := validateExpression(expr); err != nil {
		return "", nil, 0, false, err
	}
	filled, err := replaceIdentifier(body, hole, expression)
	if err != nil {
		return "", nil, 0, false, err
	}
	generated, err := generateRoute(packageName, activity, activityID, "int64", "int64", filled, preserveRoute)
	if err != nil {
		return "", nil, 0, false, err
	}
	results, passed, err := evaluateIntegerCasesContext(ctx, generated.source, activity, cases)
	return filled, results, passed, true, err
}

func bodySearchRequest(activity, activityID, body string, plan IRBodySearchPlan,
	remaining []IRBodyFillCandidate, receipt *IRBodySearchReceipt,
) (decisionroute.Request, error) {
	feedback := []bodySearchFeedback{}
	for _, attempt := range receipt.Attempts {
		observation := bodySearchFeedback{CandidateID: attempt.CandidateID,
			TypecheckPassed: attempt.TypecheckPassed, TestCasesPassed: attempt.TestCasesPassed,
			ScoringCompleted: attempt.ScoringCompleted,
			TestCasesTotal:   attempt.TestCasesTotal, AccuracyPercent: attempt.AccuracyPercent, Error: attempt.Error}
		for _, testCase := range attempt.CaseResults {
			if !testCase.Passed {
				observation.FailedCasesTotal++
				if len(observation.FailedCases) < 8 {
					observation.FailedCases = append(observation.FailedCases, testCase)
				}
			}
		}
		observation.FailedCasesTruncated = observation.FailedCasesTotal > len(observation.FailedCases)
		feedback = append(feedback, observation)
	}
	state, err := json.Marshal(struct {
		Schema            string                            `json:"schema"`
		Stage             string                            `json:"stage"`
		Activity          string                            `json:"activity"`
		ActivityID        string                            `json:"activity_id"`
		BodyIR            string                            `json:"body_ir"`
		Intent            string                            `json:"intent"`
		TrainingTestCount int                               `json:"training_test_count"`
		TrainingSuiteSHA  string                            `json:"training_suite_sha256,omitempty"`
		Remaining         []IRBodyFillCandidate             `json:"remaining_candidates"`
		PriorAttempts     []bodySearchFeedback              `json:"prior_attempts"`
		ExternalFeedback  *bodySearchExternalFeedbackPrompt `json:"external_training_feedback,omitempty"`
	}{"gooo/body-codegen-ir-search-state/v1", "choose_before_candidate_evaluation", activity, activityID,
		body, plan.Intent, len(plan.TestCases), trainingSuiteSHAForPrompt(plan, receipt), remaining, feedback,
		externalTrainingFeedbackPrompt(plan.ExternalTrainingFeedback)})
	if err != nil {
		return decisionroute.Request{}, err
	}
	options := make([]decisionroute.Option, 0, len(remaining))
	for _, candidate := range remaining {
		options = append(options, decisionroute.Option{ID: candidate.ID, Description: "Try this exact expression: " + candidate.Expression})
	}
	return decisionroute.Request{Schema: decisionroute.RequestSchema, State: string(state),
		Question: decisionroute.Question{ID: "body_ir_search", Instructions: "Choose one untried expression to fill the Gooo body hole. " +
			searchInstructions(plan.PromptProfile), Options: options},
		Fallback: remaining[0].ID, ProviderModel: plan.ProviderModel}, nil
}

func searchInstructions(profile string) string {
	if profile == "compact" {
		return "Use the declared intent and any prior failures or advisory observations. " +
			"This choice will be typechecked and tested after selection."
	}
	return "Use the declared intent and previous training failures. This choice will be typechecked and tested after selection."
}

func trainingSuiteSHAForPrompt(plan IRBodySearchPlan, receipt *IRBodySearchReceipt) string {
	if plan.PromptProfile == "compact" {
		// Compact prompt state omits the suite hash regardless of whether external feedback is attached.
		return ""
	}
	return receipt.TrainingSuiteSHA256
}

func removeBodySearchCandidate(candidates []IRBodyFillCandidate, id string) []IRBodyFillCandidate {
	remaining := make([]IRBodyFillCandidate, 0, len(candidates)-1)
	for _, candidate := range candidates {
		if candidate.ID != id {
			remaining = append(remaining, candidate)
		}
	}
	return remaining
}

func validateIRBodySearchPlan(plan IRBodySearchPlan) error {
	if plan.Schema != bodySearchPlanSchema {
		return fmt.Errorf("IR body-search plan schema must be %q", bodySearchPlanSchema)
	}
	if strings.TrimSpace(plan.Intent) == "" || utf8.RuneCountInString(plan.Intent) > 2000 || !validBodyFillIdentifier(plan.HoleID) {
		return fmt.Errorf("IR body-search intent or hole id is invalid")
	}
	if plan.CandidateGeneration == nil {
		if len(plan.Candidates) < 2 || len(plan.Candidates) > 16 || plan.MaxAttempts < 1 || plan.MaxAttempts > len(plan.Candidates) {
			return fmt.Errorf("IR body-search requires 2..16 candidates and 1..candidate-count max_attempts")
		}
	} else {
		if len(plan.Candidates) != 0 {
			return fmt.Errorf("IR body-search candidate_generation cannot be combined with a declared candidates list")
		}
		if err := validateIRBodySearchCandidateGeneration(*plan.CandidateGeneration); err != nil {
			return err
		}
		if plan.MaxAttempts < 1 || plan.MaxAttempts > plan.CandidateGeneration.MaxCandidates {
			return fmt.Errorf("IR body-search max_attempts must be 1..candidate_generation.max_candidates")
		}
	}
	if len(plan.TestCases) == 0 || len(plan.TestCases) > 4096 || len(plan.HoldoutTestCases) > 4096 {
		return fmt.Errorf("IR body-search requires 1..4096 training cases and at most 4096 holdout cases")
	}
	if err := decisionroute.ValidateProviderModel(plan.ProviderModel); err != nil {
		return err
	}
	if plan.PromptProfile != "" && plan.PromptProfile != "compact" {
		return fmt.Errorf("unsupported body-search prompt profile %q", plan.PromptProfile)
	}
	ids := map[string]bool{}
	for _, candidate := range plan.Candidates {
		if !validBodyFillIdentifier(candidate.ID) || ids[candidate.ID] || strings.TrimSpace(candidate.Expression) == "" || utf8.RuneCountInString(candidate.Expression) > 512 {
			return fmt.Errorf("IR body-search candidate %q is invalid or duplicated", candidate.ID)
		}
		ids[candidate.ID] = true
	}
	trainingInputs := map[int64]bool{}
	for _, testCase := range plan.TestCases {
		trainingInputs[testCase.Input] = true
	}
	for _, testCase := range plan.HoldoutTestCases {
		if trainingInputs[testCase.Input] {
			return fmt.Errorf("holdout input %d also appears in training cases", testCase.Input)
		}
	}
	if feedback := plan.ExternalTrainingFeedback; feedback != nil {
		if feedback.SourceDigest == "" || feedback.TrainingSuiteSHA256 == "" || feedback.CandidateID == "" {
			return fmt.Errorf("external training feedback requires source_digest, training_suite_sha256, and candidate_id")
		}
		if plan.CandidateGeneration == nil && !ids[feedback.CandidateID] {
			return fmt.Errorf("external training feedback candidate %q is not declared", feedback.CandidateID)
		}
		if len(feedback.Observations) < 1 || len(feedback.Observations) > 4096 {
			return fmt.Errorf("external training feedback requires 1..4096 observations")
		}
		cases := make(map[string]bool, len(plan.TestCases))
		for _, testCase := range plan.TestCases {
			cases[bodyFillCaseKey(testCase)] = true
		}
		seen := make(map[int64]bool, len(feedback.Observations))
		for _, observation := range feedback.Observations {
			if seen[observation.Input] {
				return fmt.Errorf("external training feedback contains duplicate input %d", observation.Input)
			}
			seen[observation.Input] = true
			if !cases[bodyFillCaseKey(IRBodyFillTestCase{Input: observation.Input, Expected: observation.Expected})] {
				return fmt.Errorf("external training observation input %d and expected value are not a declared training case", observation.Input)
			}
			if observation.Passed != (observation.Actual == observation.Expected) {
				return fmt.Errorf("external training observation passed value disagrees with actual and expected for input %d", observation.Input)
			}
		}
	}
	return nil
}

func prepareBodySearch(filename string, source []byte, activityName, holeID string) (*syntax.File, *syntax.ActivityDecl, string, string, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return nil, nil, "", "", diagnostics.Error()
	}
	if file == nil || file.Package == nil {
		return nil, nil, "", "", fmt.Errorf(".gooo source has no package declaration")
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			activity = candidate
		}
	}
	if activity == nil || !activity.ValueProgramPresent || activity.ValueProgram == "" {
		return nil, nil, "", "", fmt.Errorf("activity %q has no computes body", activityName)
	}
	if len(activity.Inputs) != 1 || activity.Inputs[0].Name != "Integer" || activity.Output != "Integer" {
		return nil, nil, "", "", fmt.Errorf("IR body search requires one Integer input and one Integer output")
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return nil, nil, "", "", err
	}
	model, err := bidir.Get(document)
	if err != nil {
		return nil, nil, "", "", err
	}
	activityID := ""
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activityName {
			activityID = string(node.ID)
		}
	}
	if activityID == "" {
		return nil, nil, "", "", fmt.Errorf("activity %q has no stable semantic identity", activityName)
	}
	body, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil || countIdentifier(body, bodyFillHoleToken(holeID)) != 1 {
		return nil, nil, "", "", fmt.Errorf("activity %q must contain exactly one body-search hole", activityName)
	}
	return file, activity, activityID, body, nil
}

func elapsedMS(started time.Time) float64 {
	return float64(time.Since(started)) / float64(time.Millisecond)
}
