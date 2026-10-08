package bodycodegen

import (
	"context"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func prepareThreeModelContext(ctx context.Context, document pathplan.Document, original *pathplan.PreparedPlan,
	model *jointdecision.ThreeModel, activityID, semanticSHA string) (*pathplan.PreparedPlan,
	*PathModelContextReceipt, bool, error) {
	text, _ := original.ThreeInput()
	declared := &PathDeclaredInputs{Text: text, SHA256: digest([]byte(text)), Bytes: len(text),
		Decisions: len(document.Plan.Decisions)}
	const schema = "gooo/compiler-three-choice-path-context/v1"
	reason := ""
	if len(document.Plan.Decisions) != 3 {
		reason = "THREE_DECISION_COUNT_UNSUPPORTED"
	} else if model.FeatureVersion() == jointdecision.RecordFieldFeatureVersion ||
		model.FeatureVersion() == jointdecision.RecordSharedFeatureVersion ||
		model.FeatureVersion() == jointdecision.RecordOriginSharedFeatureVersion ||
		model.FeatureVersion() == jointdecision.RecordGraphSharedFeatureVersion {
		reason = "FIELD_MODEL_REQUIRES_RECORD_BODY"
	}
	if reason != "" {
		r := &PathModelContextReceipt{Schema: schema, Status: "DECLINED_TO_DETERMINISTIC", ActivityID: activityID,
			SourceSemanticSHA: semanticSHA, OriginalPlanSHA: original.PlanSHA256(), MetadataSHA: model.MetadataSHA256(),
			FeatureVersion: model.FeatureVersion(), ArithmeticVersion: model.ArithmeticVersion(),
			Reason:         reason,
			DeclaredInputs: declared, Scope: "compatible three-decision input required; complete original inputs retained"}
		return original, r, true, nil
	}
	prepared, r, declined, err := prepareCompilerPathContextWithFeature(ctx, document, original, activityID,
		semanticSHA, model.MetadataSHA256(), decision.SemanticContextIntentFeatureVersion)
	if r != nil {
		r.Schema, r.FeatureVersion, r.DeclaredInputs = schema, model.FeatureVersion(), declared
		r.ArithmeticVersion = model.ArithmeticVersion()
	}
	if err != nil || declined {
		return prepared, r, declined, err
	}
	if _, err = prepared.ThreeInput(); err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_COMPLETE_INPUT_UNSUPPORTED"
		return original, r, true, nil
	}
	return prepared, r, false, nil
}
