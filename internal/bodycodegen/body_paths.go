package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const bodyPathBudget = 8 * time.Second

// BodyPathReceipt separates finite functional observations from the existing
// compiler receipt for complete lowering of a selected, typed source body.
type BodyPathReceipt struct {
	Schema                 string                  `json:"schema"`
	OriginalSourceSHA256   string                  `json:"original_source_sha256"`
	SelectedSourceSHA256   string                  `json:"selected_source_sha256,omitempty"`
	DocumentSHA256         string                  `json:"document_sha256"`
	TestSuiteSHA256        string                  `json:"test_suite_sha256"`
	SourceBaseMatched      bool                    `json:"source_base_matched"`
	SourceBinding          RouteEquivalenceReceipt `json:"source_binding"`
	Search                 pathplan.SearchResult   `json:"search"`
	NativeCases            []IRBodyFillCaseResult  `json:"native_case_results,omitempty"`
	FunctionalCompleteness float64                 `json:"finite_functional_completeness_percent"`
	Scope                  string                  `json:"scope"`
	Timing                 BodyPathTiming          `json:"timing"`
}

type BodyPathTiming struct {
	SourceBindingMS float64 `json:"source_binding_ms"`
	ModelLoadMS     float64 `json:"model_load_ms"`
	BoundedSearchMS float64 `json:"bounded_search_ms"`
	FinalEmissionMS float64 `json:"final_emission_ms"`
	TotalMS         float64 `json:"total_ms"`
	ExecutionModel  string  `json:"execution_model"`
	DecisionStage   string  `json:"decision_stage"`
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
	started := time.Now()
	receipt := &BodyPathReceipt{
		Schema: "gooo/body-codegen-typed-path-receipt/v1", OriginalSourceSHA256: digest(source),
		Scope: "declared finite cases and bounded typed alternatives; not proof of natural-language intent or all int64 inputs",
		Timing: BodyPathTiming{ExecutionModel: "single_process_source_bind_then_rank_then_finite_tdd_then_native_emit",
			DecisionStage: "all_local_predictions_before_candidate_tests_and_final_native_emission"},
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
	if err := document.Validate(); err != nil {
		return fail(err)
	}
	if document.Seed != "" && modelPath == "" {
		return fail(fmt.Errorf("typed path sampling requires an explicit local model"))
	}
	documentBytes, err := json.Marshal(document)
	if err != nil || len(documentBytes) > 128<<10 {
		return fail(fmt.Errorf("typed path document exceeds its byte budget"))
	}
	receipt.DocumentSHA256 = digest(documentBytes)
	testBytes, _ := json.Marshal(document.TestCases)
	receipt.TestSuiteSHA256 = digest(testBytes)
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() || file == nil || file.Package == nil {
		return fail(fmt.Errorf("typed path source must be a valid Gooo package"))
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			if activity != nil {
				return fail(fmt.Errorf("typed path source has duplicate activities"))
			}
			activity = candidate
		}
	}
	if activity == nil || !activity.ValueProgramPresent || len(activity.Inputs) != 1 ||
		activity.Inputs[0].Name != "Integer" || activity.Output != "Integer" || document.Plan.Base.Name != activityName {
		return fail(fmt.Errorf("typed path plan must match one source Integer -> Integer activity"))
	}
	base, err := GenerateWithPlanner(ctx, filename, source, activityName, "", "")
	if err != nil {
		return fail(err)
	}
	defaults, err := pathplan.Validate(document.Plan)
	if err != nil {
		return fail(err)
	}
	fallback, err := pathplan.Compile(document.Plan, defaults)
	if err != nil {
		return fail(err)
	}
	fallbackBody, err := rewriteLetDeclarations(fallback.GoooBody())
	if err != nil {
		return fail(err)
	}
	fallbackRoute, err := generateRoute(file.Package.Name, activityName, base.Report.ActivityID,
		"int64", "int64", fallbackBody, preserveRoute)
	if err != nil {
		return fail(err)
	}
	originalBody, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return fail(err)
	}
	receipt.SourceBinding, err = routeEquivalence(file.Package.Name, activityName, "int64", "int64",
		originalBody, fallbackRoute.source, "typed_path_fallback_matches_authoritative_source")
	if err != nil || !receipt.SourceBinding.Equivalent {
		return fail(fmt.Errorf("typed path fallback does not match the authoritative source body"))
	}
	receipt.SourceBaseMatched = true
	receipt.Timing.SourceBindingMS = elapsedMS(started)
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	var model *decision.Model
	if modelPath != "" {
		loadStarted := time.Now()
		model, err = decision.LoadPath(modelPath)
		receipt.Timing.ModelLoadMS = elapsedMS(loadStarted)
		if err != nil {
			return fail(fmt.Errorf("load explicit structural model: %w", err))
		}
	}
	searchStarted := time.Now()
	search, selected, err := pathplan.Search(ctx, document.Plan, model, document.TestCases, document.MaxAttempts, document.Seed)
	receipt.Search = search
	receipt.Timing.BoundedSearchMS = elapsedMS(searchStarted)
	if err != nil {
		return fail(err)
	}
	completed, err := replaceActivityProgram(source, activity.ValueProgramSpan, selected.GoooBody())
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
	if result.Report.ActivityID != base.Report.ActivityID {
		return fail(fmt.Errorf("typed path emission changed stable semantic identity"))
	}
	cases := make([]IRBodyFillTestCase, len(document.TestCases))
	for i, test := range document.TestCases {
		cases[i] = IRBodyFillTestCase{Input: test.Input, Expected: test.Expected}
	}
	results, passed, err := evaluateIntegerCasesContext(ctx, []byte(result.Source), activityName, cases)
	if err != nil {
		return fail(err)
	}
	// Compare all actuals, including failures, with the typed arena interpreter.
	for i, test := range document.TestCases {
		value, evaluateErr := selected.Evaluate(test.Input)
		if evaluateErr != nil || value.Int != results[i].Actual {
			return fail(fmt.Errorf("native body and typed path evaluator disagree"))
		}
	}
	if passed != search.SelectedTrainingPassed {
		return fail(fmt.Errorf("native finite pass count differs from the search receipt"))
	}
	receipt.SelectedSourceSHA256 = digest(completed)
	receipt.NativeCases = results
	receipt.FunctionalCompleteness = 100 * float64(passed) / float64(len(cases))
	receipt.Timing.FinalEmissionMS = elapsedMS(emitStarted)
	receipt.Timing.TotalMS = elapsedMS(started)
	result.Report.BodyPaths = receipt
	return result, nil
}
