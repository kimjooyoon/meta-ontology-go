package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/orderjudge"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const bodyPathBudget = 8 * time.Second

// BodyPathReceipt separates finite functional observations from the existing
// compiler receipt for complete lowering of a selected, typed source body.
type BodyPathReceipt struct {
	Schema                 string                       `json:"schema"`
	OriginalSourceSHA256   string                       `json:"original_source_sha256"`
	SelectedSourceSHA256   string                       `json:"selected_source_sha256,omitempty"`
	SourceFormat           string                       `json:"source_format,omitempty"`
	DocumentSHA256         string                       `json:"document_sha256"`
	TestSuiteSHA256        string                       `json:"test_suite_sha256"`
	SearchConfigSHA256     string                       `json:"search_config_sha256"`
	SearchConfig           json.RawMessage              `json:"search_config"`
	DeclaredTestCases      int                          `json:"declared_test_cases"`
	LocalModelRequested    bool                         `json:"local_model_requested"`
	SearchStarted          bool                         `json:"search_started"`
	SourceBaseMatched      bool                         `json:"source_base_matched"`
	SourceBinding          RouteEquivalenceReceipt      `json:"source_binding"`
	Search                 pathplan.SearchResult        `json:"search"`
	OrderJudgment          *orderjudge.SearchReceipt    `json:"whole_candidate_judgment,omitempty"`
	OrderPreparation       *OrderPreparationReceipt     `json:"whole_candidate_preparation,omitempty"`
	Progress               []pathplan.SessionProgress   `json:"session_progress,omitempty"`
	Feedback               []pathplan.FeedbackReceipt   `json:"feedback_judgments,omitempty"`
	ConditionProgress      []pathplan.ConditionProgress `json:"condition_session_progress,omitempty"`
	ConditionFeedback      []pathplan.ConditionRanking  `json:"condition_feedback_judgments,omitempty"`
	FeedbackUnfixed        bool                         `json:"feedback_unfixed,omitempty"`
	ModelRetention         *RetainedModelInfo           `json:"model_retention,omitempty"`
	ModelContext           *PathModelContextReceipt     `json:"model_context,omitempty"`
	Diagnosis              *pathplan.Diagnosis          `json:"diagnosis,omitempty"`
	DiagnosisOptionsSHA256 string                       `json:"diagnosis_options_sha256,omitempty"`
	DiagnosisBudget        int                          `json:"diagnosis_budget,omitempty"`
	DiagnosisScope         string                       `json:"diagnosis_scope,omitempty"`
	Observation            *PathObservationReceipt      `json:"observation,omitempty"`
	Resolution             *PathResolutionReceipt       `json:"resolution,omitempty"`
	NativeCases            []IRBodyFillCaseResult       `json:"native_case_results,omitempty"`
	Conditions             *PathConditionReceipt        `json:"conditions,omitempty"`
	FunctionalCompleteness float64                      `json:"finite_functional_completeness_percent"`
	Scope                  string                       `json:"scope"`
	Timing                 BodyPathTiming               `json:"timing"`
}

type BodyPathTiming struct {
	PlanPrepareMS    float64 `json:"plan_prepare_ms"`
	SourceBindingMS  float64 `json:"source_binding_ms"`
	ModelLoadMS      float64 `json:"model_load_ms"`
	ContextPrepareMS float64 `json:"context_prepare_ms,omitempty"`
	BoundedSearchMS  float64 `json:"bounded_search_ms"`
	DiagnosisMS      float64 `json:"diagnosis_ms,omitempty"`
	ObservationMS    float64 `json:"observation_ms,omitempty"`
	ResolutionMS     float64 `json:"resolution_ms,omitempty"`
	FinalEmissionMS  float64 `json:"final_emission_ms"`
	TotalMS          float64 `json:"total_ms"`
	ExecutionModel   string  `json:"execution_model"`
	DecisionStage    string  `json:"decision_stage"`
}

type BodyPathError struct {
	Receipt *BodyPathReceipt
	Cause   error
}

func (failure *BodyPathError) Error() string { return failure.Cause.Error() }
func (failure *BodyPathError) Unwrap() error { return failure.Cause }

// GenerateWithTypedPaths is an explicit, local experiment. The finite contract
// authorizes edits only at the declared typed paths after the original body is
// bound to the fallback plan. It never changes source files or calls a network
// provider. Model ranking precedes candidate tests and final native emission.
func GenerateWithTypedPaths(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string) (Result, error) {
	return generateWithTypedPathBatches(ctx, filename, source, activityName, document, modelPath, 0, nil)
}

// GenerateWithTypedPathBatches ranks once and advances new typed candidates in
// batches. The document's total 1..64 attempt budget and source authority remain.
func GenerateWithTypedPathBatches(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string, stepAttempts int) (Result, error) {
	if stepAttempts < 1 || stepAttempts > 64 {
		return Result{}, fmt.Errorf("typed path step attempts must be 1..64")
	}
	return generateWithTypedPathBatches(ctx, filename, source, activityName, document, modelPath, stepAttempts, nil)
}

type typedPathFeedback struct {
	rounds  int
	ci      *pathplan.CIHint
	unfixed bool
}

// GenerateWithTypedPathFeedback explicitly reconsiders remaining paths using
// the original local model and observed finite failures. CI is caller context.
func GenerateWithTypedPathFeedback(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string, stepAttempts, rounds int, ci *pathplan.CIHint) (Result, error) {
	return generateTypedFeedback(ctx, filename, source, activityName, document, modelPath, stepAttempts, rounds, ci, false)
}

// GenerateWithTypedPathUnfixedFeedback skips predictions only for coordinates
// constant across all unattempted masks. Ordinary source and finite checks remain.
func GenerateWithTypedPathUnfixedFeedback(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string, stepAttempts, rounds int, ci *pathplan.CIHint) (Result, error) {
	return generateTypedFeedback(ctx, filename, source, activityName, document, modelPath, stepAttempts, rounds, ci, true)
}

func generateTypedFeedback(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string, stepAttempts, rounds int, ci *pathplan.CIHint, unfixed bool) (Result, error) {
	if modelPath == "" || stepAttempts < 1 || stepAttempts > 64 || rounds < 1 || rounds > 16 {
		return Result{}, fmt.Errorf("typed path feedback requires model, 1..64 step and 1..16 rounds")
	}
	if err := ci.Validate(); err != nil {
		return Result{}, err
	}
	return generateWithTypedPathBatches(ctx, filename, source, activityName, document, modelPath, stepAttempts,
		&typedPathFeedback{rounds: rounds, ci: ci, unfixed: unfixed})
}

func generateWithTypedPathBatches(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, modelPath string, stepAttempts int, feedback *typedPathFeedback) (Result, error) {
	return generateTypedPathRequest(ctx, filename, source, activityName, document,
		typedPathModel{path: modelPath}, stepAttempts, feedback)
}

func generateTypedPathRequest(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, models typedPathModel, stepAttempts int, feedback *typedPathFeedback) (Result, error) {
	started := time.Now()
	searchConfig := typedPathSearchConfig(stepAttempts, feedback)
	receipt := &BodyPathReceipt{
		Schema: "gooo/body-codegen-typed-path-receipt/v1", OriginalSourceSHA256: digest(source),
		LocalModelRequested: models.path != "" || models.model != nil || models.joint != nil || models.three != nil || models.order != nil || models.condition != nil,
		SearchConfigSHA256:  digest(searchConfig), SearchConfig: searchConfig,
		Scope: "declared finite cases and bounded typed alternatives; not proof of natural-language intent or all int64 inputs",
		Timing: BodyPathTiming{ExecutionModel: "single_process_source_bind_then_rank_then_finite_tdd_then_native_emit",
			DecisionStage: "all_local_predictions_before_candidate_tests_and_final_native_emission"},
	}
	if models.retention != nil {
		info := *models.retention
		receipt.ModelRetention = &info
	}
	if feedback != nil {
		receipt.Timing.ExecutionModel = "single_process_source_bind_then_rank_then_interleaved_finite_tdd_feedback_then_native_emit"
		receipt.Timing.DecisionStage = "local_initial_ranking_and_explicit_partial_batch_feedback_before_final_native_emission"
		receipt.FeedbackUnfixed = feedback.unfixed
	}
	fail := func(err error) (Result, error) {
		receipt.Timing.TotalMS = elapsedMS(started)
		return Result{}, &BodyPathError{Receipt: receipt, Cause: err}
	}
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 {
		return fail(fmt.Errorf("typed path codegen requires a context and source of at most 128 KiB"))
	}
	ctx, cancel := context.WithTimeout(ctx, bodyPathBudget)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	prepareStarted := time.Now()
	if err := bindTypedPathDocument(document, receipt); err != nil {
		return fail(err)
	}
	prepared, err := document.Prepare()
	receipt.Timing.PlanPrepareMS = elapsedMS(prepareStarted)
	if err != nil {
		return fail(err)
	}
	if document.Seed != "" && models.path == "" && models.model == nil && models.joint == nil && models.three == nil && models.order == nil && models.condition == nil {
		return fail(fmt.Errorf("typed path sampling requires an explicit local model"))
	}
	bound, err := bindTypedPathSource(ctx, filename, source, activityName, prepared, receipt)
	if err != nil {
		return fail(err)
	}
	base := bound.base
	effectiveCases, err := observeTypedPaths(ctx, filename, source, activityName, prepared,
		document.TestCases, models.observation, receipt)
	if err != nil {
		return fail(err)
	}
	// Model context remains tied to the complete original document. Only the
	// separately recorded finite evaluation suite gains oracle observations.
	searchDocument := document
	searchDocument.TestCases = effectiveCases
	resolved, err := resolveObservedPath(ctx, document, prepared, receipt)
	if err != nil {
		return fail(err)
	}
	if resolved != nil {
		if err := diagnoseSelectedPath(ctx, prepared, searchDocument, models.diagnosis, receipt); err != nil {
			return fail(err)
		}
		return emitSelectedTypedPath(ctx, filename, source, activityName, bound, resolved, effectiveCases, receipt, started)
	}
	model := models.model
	joint := models.joint
	three := models.three
	order := models.order
	condition := models.condition
	if models.path != "" {
		loadStarted := time.Now()
		var loaded typedPathModel
		loaded, err = loadTypedStructuralModel(models.path)
		model, joint, three = loaded.model, loaded.joint, loaded.three
		order = loaded.order
		condition = loaded.condition
		receipt.Timing.ModelLoadMS = elapsedMS(loadStarted)
		if err != nil {
			return fail(fmt.Errorf("load explicit structural model: %w", err))
		}
	}
	if order != nil {
		if document.Seed != "" || stepAttempts != 0 || feedback != nil {
			return fail(fmt.Errorf("whole-candidate judge requires unseeded search without batch or feedback options"))
		}
		return generateOrderTypedPath(ctx, filename, source, activityName, searchDocument, prepared,
			bound, order, models.diagnosis, receipt, started)
	}
	contextStarted := time.Now()
	var contextDeclined bool
	if condition != nil {
		if feedback != nil && (feedback.ci != nil || feedback.unfixed) {
			return fail(fmt.Errorf("candidate model requires its declared observation channels; CI hints and unfixed feedback are outside its feature contract"))
		}
		prepared, receipt.ModelContext, contextDeclined, err = prepareConditionModelContext(ctx, document, prepared, condition, base.Report.ActivityID, receipt.SourceBinding.SourceSemanticDigest)
	} else if three != nil {
		prepared, receipt.ModelContext, contextDeclined, err = prepareThreeModelContext(ctx, document, prepared,
			three, base.Report.ActivityID, receipt.SourceBinding.SourceSemanticDigest)
	} else if joint != nil {
		prepared, receipt.ModelContext, contextDeclined, err = prepareJointModelContext(ctx, document, prepared, joint, base.Report.ActivityID, receipt.SourceBinding.SourceSemanticDigest)
	} else {
		prepared, receipt.ModelContext, contextDeclined, err = preparePathModelContext(ctx, document, prepared, model,
			base.Report.ActivityID, receipt.SourceBinding.SourceSemanticDigest)
	}
	if receipt.ModelContext != nil {
		receipt.Timing.ContextPrepareMS = elapsedMS(contextStarted)
	}
	if err != nil {
		return fail(err)
	}
	searchSeed := document.Seed
	if contextDeclined {
		receipt.ModelContext.SeedSkipped, receipt.ModelContext.FeedbackSkipped = searchSeed != "", feedback != nil
		model, feedback, searchSeed = nil, nil, ""
		joint = nil
		three = nil
		condition = nil
		receipt.FeedbackUnfixed = false
		receipt.Timing.ExecutionModel = "source_bind_then_context_decline_then_deterministic_finite_tdd_then_native_emit"
		receipt.Timing.DecisionStage = "no_predictions_representation_declined"
	}
	searchStarted := time.Now()
	if models.observation != nil {
		receipt.Timing.ExecutionModel = "source_bind_then_oracle_observations_then_model_rank_finite_tdd_then_native_emit"
		receipt.Timing.DecisionStage = "local_ranking_after_recorded_oracle_observations; original_model_input_preserved"
		if contextDeclined {
			receipt.Timing.DecisionStage = "no_predictions_representation_declined; recorded_oracle_observations_retained"
		}
	}
	receipt.SearchStarted = true
	var search pathplan.SearchResult
	var selected *bodyplan.Program
	if condition != nil {
		rounds := 0
		if feedback != nil {
			rounds = feedback.rounds
		}
		search, selected, receipt.ConditionProgress, receipt.ConditionFeedback, err = condition.search(ctx, prepared,
			effectiveCases, document.MaxAttempts, min(document.MaxAttempts, max(1, stepAttempts)), searchSeed, rounds)
	} else if three != nil {
		rounds := 0
		var ci *pathplan.CIHint
		if feedback != nil {
			rounds, ci = feedback.rounds, feedback.ci
		}
		search, selected, receipt.Progress, receipt.Feedback, err = prepared.SearchThreeFeedbackBatches(ctx, three,
			effectiveCases, document.MaxAttempts, max(1, stepAttempts), searchSeed, rounds, ci)
	} else if joint != nil {
		rounds := 0
		var ci *pathplan.CIHint
		if feedback != nil {
			rounds, ci = feedback.rounds, feedback.ci
		}
		search, selected, receipt.Progress, receipt.Feedback, err = prepared.SearchJointFeedbackBatches(ctx, joint, effectiveCases, document.MaxAttempts, max(1, stepAttempts), searchSeed, rounds, ci)
	} else if feedback != nil && feedback.unfixed {
		search, selected, receipt.Progress, receipt.Feedback, err = prepared.SearchFeedbackBatchesUnfixed(ctx, model,
			effectiveCases, document.MaxAttempts, stepAttempts, searchSeed, feedback.rounds, feedback.ci)
	} else if feedback != nil {
		search, selected, receipt.Progress, receipt.Feedback, err = prepared.SearchFeedbackBatches(ctx, model, effectiveCases,
			document.MaxAttempts, stepAttempts, searchSeed, feedback.rounds, feedback.ci)
	} else if stepAttempts == 0 {
		search, selected, err = prepared.Search(ctx, model, effectiveCases, document.MaxAttempts, searchSeed)
	} else {
		search, selected, receipt.Progress, err = prepared.SearchBatches(ctx, model, effectiveCases,
			document.MaxAttempts, stepAttempts, searchSeed)
	}
	receipt.Search = search
	receipt.Timing.BoundedSearchMS = elapsedMS(searchStarted)
	if err != nil {
		return fail(err)
	}
	if err := diagnoseSelectedPath(ctx, prepared, searchDocument, models.diagnosis, receipt); err != nil {
		return fail(err)
	}
	return emitSelectedTypedPath(ctx, filename, source, activityName, bound, selected, effectiveCases, receipt, started)
}

func emitSelectedTypedPath(ctx context.Context, filename string, source []byte, activityName string,
	bound typedPathSource, selected *bodyplan.Program, effectiveCases []pathplan.TestCase,
	receipt *BodyPathReceipt, started time.Time) (Result, error) {
	fail := func(err error) (Result, error) {
		receipt.Timing.TotalMS = elapsedMS(started)
		return Result{}, &BodyPathError{Receipt: receipt, Cause: err}
	}
	completed, err := selectedTypedPathSource(source, bound.activity, selected.GoooBody(), bodyPathSelection(receipt).Choices,
		bound.activity.Assembly != nil)
	if err != nil {
		return fail(err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	emitStarted := time.Now()
	result, err := GenerateWithPlanner(ctx, filename, completed, activityName, "", "")
	if err != nil {
		return fail(err)
	}
	if result.Report.ActivityID != bound.base.Report.ActivityID {
		return fail(fmt.Errorf("typed path emission changed stable semantic identity"))
	}
	conditions, err := selectedPathConditions(ctx, bound.prepared, bodyPathSelection(receipt).Choices,
		selected, activityName, []byte(result.Source))
	if conditions != nil {
		receipt.Conditions = conditions
	}
	if err != nil {
		return fail(err)
	}
	if conditions != nil {
		if !sameConditionResults(conditions.Results, bodyPathSelection(receipt).Conditions) {
			return fail(fmt.Errorf("selected condition observations differ from source-bound search"))
		}
		receipt.Conditions = conditions
	}
	cases := make([]IRBodyFillTestCase, len(effectiveCases))
	for i, test := range effectiveCases {
		cases[i] = IRBodyFillTestCase{Input: test.Input, Expected: test.Expected}
	}
	results, passed, err := evaluateIntegerCasesContext(ctx, []byte(result.Source), activityName, cases)
	if err != nil {
		return fail(err)
	}
	// Compare all actuals, including failures, with the typed arena interpreter.
	for i, test := range effectiveCases {
		value, evaluateErr := selected.Evaluate(test.Input)
		if evaluateErr != nil || value.Int != results[i].Actual {
			return fail(fmt.Errorf("native body and typed path evaluator disagree"))
		}
	}
	expectedPassed, _ := bodyPathSelectedScore(receipt)
	if passed != expectedPassed {
		return fail(fmt.Errorf("native finite pass count differs from the selection record"))
	}
	receipt.SelectedSourceSHA256 = digest(completed)
	if bound.activity.Assembly != nil {
		receipt.SourceFormat = sourceAssemblyCheckpointFormat
		result.GoooSource = string(completed)
	}
	receipt.NativeCases = results
	receipt.FunctionalCompleteness = 100 * float64(passed) / float64(len(cases))
	receipt.Timing.FinalEmissionMS = elapsedMS(emitStarted)
	receipt.Timing.TotalMS = elapsedMS(started)
	result.Report.BodyPaths = receipt
	populateCompletenessReceipt(&result.Report, "")
	return result, nil
}
