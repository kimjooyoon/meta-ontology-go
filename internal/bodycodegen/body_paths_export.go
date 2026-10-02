package bodycodegen

import (
	"context"
	"fmt"
	"time"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Context text is disclosed only by this explicit export API, not normal receipts.
type ExportedPathInput struct {
	PathContextInput
	Text string `json:"text"`
}

type TypedPathContextExport struct {
	Schema               string                   `json:"schema"`
	OriginalSourceSHA256 string                   `json:"original_source_sha256"`
	DocumentSHA256       string                   `json:"document_sha256"`
	TestSuiteSHA256      string                   `json:"test_suite_sha256"`
	SourceBinding        RouteEquivalenceReceipt  `json:"source_binding"`
	Context              *PathModelContextReceipt `json:"context"`
	Inputs               []ExportedPathInput      `json:"inputs"`
	Timing               BodyPathTiming           `json:"timing"`
	ModelPredictions     int                      `json:"model_predictions"`
	CandidateTests       int                      `json:"candidate_tests"`
	SelectedEmission     bool                     `json:"selected_emission"`
	RepositoryWrites     int                      `json:"repository_writes"`
	Scope                string                   `json:"scope"`
}

// ExportTypedPathContext uses the same source binding and encoder as ranking.
// It emits/typechecks original/fallback validation projections. It does not
// load a model, test candidates, or emit a selected candidate. Text can contain
// caller data; callers choose what to export or publish. A decline exports no text.
func ExportTypedPathContext(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document) (TypedPathContextExport, error) {
	return ExportTypedPathContextWithFeature(ctx, filename, source, activityName, document,
		decision.SplitContextIntentFeatureVersion)
}

// ExportTypedPathContextWithFeature explicitly selects a compiler input ABI.
// The default API/CLI continue to use v2; v3 weights must opt in via metadata.
func ExportTypedPathContextWithFeature(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, featureVersion string) (TypedPathContextExport, error) {
	started := time.Now()
	receipt := &BodyPathReceipt{Schema: "gooo/body-context-export-validation/v1",
		OriginalSourceSHA256: digest(source), Timing: BodyPathTiming{
			ExecutionModel: "source_bind_then_compiler_context_export",
			DecisionStage:  "no_model_no_candidate_tests_no_selected_emission"}}
	fail := func(err error) (TypedPathContextExport, error) {
		return pathContextExportFailure(receipt, started, err)
	}
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 {
		return fail(fmt.Errorf("typed path context requires a context and source of at most 128 KiB"))
	}
	if featureVersion != decision.SplitContextIntentFeatureVersion && featureVersion != decision.SemanticContextIntentFeatureVersion {
		return fail(fmt.Errorf("unsupported compiler context feature version"))
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
	bound, err := bindTypedPathSource(ctx, filename, source, activityName, prepared, receipt)
	if err != nil {
		return fail(err)
	}
	return exportBoundPathContext(ctx, document, prepared, bound.base.Report.ActivityID, featureVersion, receipt, started)
}

func pathContextExportFailure(receipt *BodyPathReceipt, started time.Time, err error) (TypedPathContextExport, error) {
	receipt.Timing.TotalMS = elapsedMS(started)
	return TypedPathContextExport{}, &BodyPathError{Receipt: receipt, Cause: err}
}

func exportBoundPathContext(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	activityID, featureVersion string, receipt *BodyPathReceipt, started time.Time) (TypedPathContextExport, error) {
	contextStarted := time.Now()
	_, modelContext, declined, err := prepareCompilerPathContextWithFeature(ctx, document, prepared,
		activityID, receipt.SourceBinding.SourceSemanticDigest, "", featureVersion)
	receipt.ModelContext = modelContext
	receipt.Timing.ContextPrepareMS = elapsedMS(contextStarted)
	if err != nil {
		return pathContextExportFailure(receipt, started, err)
	}
	inputs, err := exportedPathInputs(document, prepared, modelContext, declined, featureVersion)
	if err != nil {
		return pathContextExportFailure(receipt, started, err)
	}
	receipt.Timing.TotalMS = elapsedMS(started)
	schema := "gooo/compiler-path-input-export/v1"
	if featureVersion == decision.SemanticContextIntentFeatureVersion {
		schema = "gooo/compiler-path-input-export/v2"
	}
	return TypedPathContextExport{Schema: schema,
		OriginalSourceSHA256: receipt.OriginalSourceSHA256, DocumentSHA256: receipt.DocumentSHA256,
		TestSuiteSHA256: receipt.TestSuiteSHA256, SourceBinding: receipt.SourceBinding,
		Context: modelContext, Inputs: inputs, Timing: receipt.Timing,
		Scope: "explicit source-bound model inputs; validation projections generated; no model predictions, candidate tests, selected emission, or writes; seed unused; finite expectations excluded from input text"}, nil
}

func exportedPathInputs(document pathplan.Document, prepared *pathplan.PreparedPlan,
	modelContext *PathModelContextReceipt, declined bool, featureVersion string) ([]ExportedPathInput, error) {
	inputs := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	if declined {
		return inputs, nil
	}
	var facts bodyplan.Plan
	if featureVersion == decision.SplitContextIntentFeatureVersion {
		facts = fallbackContextFacts(document.Plan)
	}
	for i, choice := range document.Plan.Decisions {
		text, input, reason := encodeCompilerPathContext(featureVersion, facts, prepared, choice)
		if reason != "" || input != modelContext.Inputs[i] {
			return nil, fmt.Errorf("context export differs from ranking input")
		}
		inputs = append(inputs, ExportedPathInput{PathContextInput: input, Text: text})
	}
	return inputs, nil
}
