package bodycodegen

import (
	"context"
	"fmt"
	"time"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

type PathModelCompatibility struct {
	Schema string            `json:"schema"`
	Status string            `json:"status"`
	Reason string            `json:"reason"`
	Model  RetainedModelInfo `json:"model"`
	Scope  string            `json:"scope"`
}

// Joint heads consume this complete framing, not individual choice strings.
type CompletePathModelInput struct {
	Text   string `json:"text"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

// Inspection uses construction's existing model-specific preparation functions.
func (m typedPathModel) prepareContext(ctx context.Context, document pathplan.Document,
	original *pathplan.PreparedPlan, activityID, semanticSHA string) (
	*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	if m.condition != nil {
		return prepareConditionModelContext(ctx, document, original, m.condition, activityID, semanticSHA)
	}
	if m.three != nil {
		return prepareThreeModelContext(ctx, document, original, m.three, activityID, semanticSHA)
	}
	if m.joint != nil {
		return prepareJointModelContext(ctx, document, original, m.joint, activityID, semanticSHA)
	}
	return preparePathModelContext(ctx, document, original, m.model, activityID, semanticSHA)
}

// ExportTypedPathModelContext binds source before loading one explicit model.
// It uses construction's encoders without predictions, candidate tests or writes.
// A representation decline is retained as a successful inspection result.
func ExportTypedPathModelContext(ctx context.Context, filename string, source []byte, activity string,
	document pathplan.Document, modelPath, explicitFeature string) (TypedPathContextExport, error) {
	if modelPath == "" {
		return TypedPathContextExport{}, fmt.Errorf("explicit local path model required")
	}
	return exportTypedPathContext(ctx, filename, source, activity, document, explicitFeature, modelPath)
}

func exportBoundPathModelContext(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	activityID, explicitFeature, modelPath string, receipt *BodyPathReceipt, started time.Time) (TypedPathContextExport, error) {
	fail := func(err error) (TypedPathContextExport, error) {
		return pathContextExportFailure(receipt, started, err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	loadStarted := time.Now()
	receipt.Timing.ExecutionModel = "source_bind_then_explicit_model_load_then_context_export"
	receipt.Timing.DecisionStage = "model_load_attempt_without_predictions_or_candidate_tests_or_selected_emission"
	g, err := NewTypedPathGenerator(modelPath)
	receipt.Timing.ModelLoadMS = elapsedMS(loadStarted)
	if err != nil {
		return fail(err)
	}
	receipt.Timing.DecisionStage = "model_loaded_without_predictions_or_candidate_tests_or_selected_emission"
	if g.order != nil {
		return fail(fmt.Errorf("body-context model preflight does not yet cover whole-candidate judges"))
	}
	info := g.Info()
	if explicitFeature != "" && explicitFeature != info.FeatureVersion {
		return fail(fmt.Errorf("explicit feature version differs from loaded model feature %q", info.FeatureVersion))
	}
	contextStarted := time.Now()
	models := typedPathModel{model: g.model, joint: g.joint, three: g.three, condition: g.condition}
	ranked, modelContext, declined, err := models.prepareContext(ctx, document, prepared,
		activityID, receipt.SourceBinding.SourceSemanticDigest)
	if err != nil {
		return fail(err)
	}
	inputs, err := modelContextInputs(document, prepared, g, &modelContext, declined,
		activityID, receipt.SourceBinding.SourceSemanticDigest)
	receipt.ModelContext = modelContext
	receipt.Timing.ContextPrepareMS = elapsedMS(contextStarted)
	if err != nil {
		return fail(err)
	}
	return finishPathModelContext(ctx, ranked, g, modelContext, inputs, declined, receipt, started)
}

func modelContextInputs(document pathplan.Document, prepared *pathplan.PreparedPlan, g *TypedPathGenerator,
	modelContext **PathModelContextReceipt, declined bool, activityID, semanticSHA string) ([]ExportedPathInput, error) {
	if declined {
		return []ExportedPathInput{}, nil
	}
	if g.condition != nil {
		return exportedConditionInputs(document, prepared)
	}
	if *modelContext != nil {
		feature := decision.SemanticContextIntentFeatureVersion
		if g.model != nil {
			feature = g.model.FeatureVersion()
		}
		return exportedPathInputs(document, prepared, *modelContext, false, feature)
	}
	// Legacy per-choice models consume original caller text exactly as declared.
	r := &PathModelContextReceipt{Schema: "gooo/compiler-declared-path-context/v1", Status: "ENCODED",
		ActivityID: activityID, SourceSemanticSHA: semanticSHA, OriginalPlanSHA: prepared.PlanSHA256(),
		RankedPlanSHA: prepared.PlanSHA256(), MetadataSHA: g.info.MetadataSHA256, FeatureVersion: g.info.FeatureVersion,
		Inputs: []PathContextInput{}, Scope: "source-bound original caller text; legacy model input is not compiler-projected context"}
	inputs := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	var features [decision.FeatureDim]float32
	for _, choice := range document.Plan.Decisions {
		if err := g.model.FeaturesInto(choice.Intent, &features); err != nil {
			return nil, err
		}
		sha := digest([]byte(choice.Intent))
		input := PathContextInput{DecisionID: choice.ID, OriginalIntentSHA: sha, NaturalIntentSHA: sha,
			InputSHA: sha, Bytes: len(choice.Intent)}
		r.Inputs = append(r.Inputs, input)
		inputs = append(inputs, ExportedPathInput{PathContextInput: input, Text: choice.Intent})
	}
	*modelContext = r
	return inputs, nil
}

func finishPathModelContext(ctx context.Context, ranked *pathplan.PreparedPlan, g *TypedPathGenerator,
	modelContext *PathModelContextReceipt, inputs []ExportedPathInput, declined bool,
	receipt *BodyPathReceipt, started time.Time) (TypedPathContextExport, error) {
	var complete *CompletePathModelInput
	if !declined && (g.joint != nil || g.three != nil) {
		var text string
		var err error
		if g.three != nil {
			text, err = ranked.ThreeInput()
		} else {
			text, err = ranked.JointInput()
		}
		if err != nil {
			return pathContextExportFailure(receipt, started, err)
		}
		complete = &CompletePathModelInput{Text: text, SHA256: digest([]byte(text)), Bytes: len(text)}
	}
	if err := ctx.Err(); err != nil {
		return pathContextExportFailure(receipt, started, err)
	}
	status, reason := modelContext.Status, modelContext.Reason
	if status == "ENCODED" {
		status, reason = "READY_FOR_RANKING", "SOURCE_INPUT_ENCODED"
	}
	receipt.Timing.TotalMS = elapsedMS(started)
	info := g.Info()
	info.Scope = "one preflight model load; setup included in model_load_ms and total_ms; no inference"
	return TypedPathContextExport{Schema: "gooo/compiler-path-model-input-export/v1",
		OriginalSourceSHA256: receipt.OriginalSourceSHA256, DocumentSHA256: receipt.DocumentSHA256,
		TestSuiteSHA256: receipt.TestSuiteSHA256, SourceBinding: receipt.SourceBinding,
		Context: modelContext, Inputs: inputs, CompleteModelInput: complete, Timing: receipt.Timing,
		ModelCompatibility: &PathModelCompatibility{Schema: "gooo/path-model-compatibility/v1", Status: status,
			Reason: reason, Model: info, Scope: "verified artifact and source input representation; ranking and correctness remain unmeasured"},
		Scope: "source-bound model input inspection; zero predictions, candidate tests, selected emissions and repository writes; seed and outcomes unused"}, nil
}
