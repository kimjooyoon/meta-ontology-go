package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/executiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/orderjudge"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Dispatch reads bounded metadata; each loader enforces its closed artifact
// contract. Fresh source binding remains part of each construction request.
func loadTypedStructuralModel(name string) (typedPathModel, error) {
	raw, err := readStructuralModelBytes(name)
	if err != nil {
		return typedPathModel{}, err
	}
	if err = decision.RejectDuplicateJSONKeys(raw); err != nil {
		return typedPathModel{}, err
	}
	var selector struct {
		Schema  string `json:"schema"`
		Feature string `json:"feature_version"`
	}
	if err = json.Unmarshal(raw, &selector); err != nil {
		return typedPathModel{}, err
	}
	if selector.Schema == conditiondecision.Schema || selector.Schema == executiondecision.Schema || selector.Schema == flowdecision.Schema {
		return decodeCandidateModel(raw, selector.Schema)
	}
	if len(raw) > 64<<10 {
		return typedPathModel{}, fmt.Errorf("bounded regular structural metadata required")
	}
	if selector.Schema == orderjudge.Schema {
		model, err := loadOrderJudge(name, raw)
		return typedPathModel{order: model}, err
	}
	if selector.Schema == jointdecision.ThreeSchema {
		if selector.Feature == jointdecision.RecordFieldFeatureVersion {
			model, err := jointdecision.LoadRecordThree(name)
			return typedPathModel{three: model}, err
		}
		if selector.Feature == jointdecision.ThreeBagFeatureVersion {
			model, err := jointdecision.LoadThreeBag(name)
			return typedPathModel{three: model}, err
		}
		model, err := jointdecision.LoadThree(name)
		return typedPathModel{three: model}, err
	}
	if selector.Schema == jointdecision.SharedThreeSchema {
		return loadCompilerSharedThree(name, selector.Feature)
	}
	if selector.Schema == jointdecision.Schema {
		model, err := jointdecision.Load(name)
		return typedPathModel{joint: model}, err
	}
	if selector.Schema != decision.PathMetadataSchema {
		return typedPathModel{}, fmt.Errorf("explicit structural model schema required")
	}
	model, err := decision.LoadPath(name)
	return typedPathModel{model: model}, err
}
func prepareJointModelContext(ctx context.Context, document pathplan.Document, original *pathplan.PreparedPlan,
	model *jointdecision.Model, activityID, semanticSHA string) (*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	if len(document.Plan.Decisions) != 2 {
		r := &PathModelContextReceipt{Schema: "gooo/compiler-joint-path-context/v1", Status: "DECLINED_TO_DETERMINISTIC", ActivityID: activityID, SourceSemanticSHA: semanticSHA, OriginalPlanSHA: original.PlanSHA256(), MetadataSHA: model.MetadataSHA256(), FeatureVersion: jointdecision.FeatureVersion, Reason: "JOINT_DECISION_COUNT_UNSUPPORTED", Scope: "Joint head requires exactly two typed decisions; retained deterministic continuation."}
		return original, r, true, nil
	}
	prepared, r, declined, err := prepareCompilerPathContextWithFeature(ctx, document, original, activityID, semanticSHA, model.MetadataSHA256(), decision.SemanticContextIntentFeatureVersion)
	if r != nil {
		r.Schema, r.FeatureVersion = "gooo/compiler-joint-path-context/v1", jointdecision.FeatureVersion
	}
	if err != nil || declined {
		return prepared, r, declined, err
	}
	if _, err = prepared.JointInput(); err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "JOINT_COMPLETE_INPUT_UNSUPPORTED"
		return original, r, true, nil
	}
	return prepared, r, false, nil
}
