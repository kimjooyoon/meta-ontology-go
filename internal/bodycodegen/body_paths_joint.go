package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"io"
	"os"
)

// Schema dispatch occurs after source binding; each loader enforces its full
// closed metadata and weights contract. The operation classifier is rejected.
func loadTypedStructuralModel(name string) (*decision.Model, *jointdecision.Model, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > 64<<10 {
		return nil, nil, fmt.Errorf("bounded regular structural metadata required")
	}
	raw, err := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if err != nil {
		return nil, nil, err
	}
	if len(raw) > 64<<10 {
		return nil, nil, fmt.Errorf("structural metadata grew beyond bound")
	}
	if err = decision.RejectDuplicateJSONKeys(raw); err != nil {
		return nil, nil, err
	}
	var selector struct {
		Schema string `json:"schema"`
	}
	if err = json.Unmarshal(raw, &selector); err != nil {
		return nil, nil, err
	}
	if selector.Schema == jointdecision.Schema {
		model, err := jointdecision.Load(name)
		return nil, model, err
	}
	if selector.Schema != decision.PathMetadataSchema {
		return nil, nil, fmt.Errorf("explicit structural model schema required")
	}
	model, err := decision.LoadPath(name)
	return model, nil, err
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
