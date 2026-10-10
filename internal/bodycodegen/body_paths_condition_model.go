package bodycodegen

import (
	"context"
	"encoding/binary"
	"math"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/executiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const conditionModelSchema = conditiondecision.Schema
const conditionModelTensorBytes = conditiondecision.ParameterCount * 4

// The raw artifact identity is distinct from its formatting-independent model
// fingerprint. Both bind the same immutable, once-decoded weight arrays.
type conditionPathModel struct {
	model       *conditiondecision.Model
	execution   *executiondecision.Model
	flow        *flowdecision.Model
	contract    declaredContractModel
	artifactSHA string
}

func conditionFeatureDigest(features [decision.FeatureDim]float32) string {
	return candidateFeatureDigest(features[:])
}

func candidateFeatureDigest(features []float32) string {
	var raw [contractdecision.OrderedFeatureDim * 4]byte
	for i, value := range features {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(value))
	}
	return digest(raw[:len(features)*4])
}

func prepareConditionModelContext(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	model *conditionPathModel, activityID, semanticSHA string) (*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	if model.contract != nil {
		return prepareContractModelContext(ctx, document, prepared, model, activityID, semanticSHA)
	}
	r := &PathModelContextReceipt{Schema: "gooo/compiler-condition-model-context/v1", Status: "ENCODED", ActivityID: activityID,
		SourceSemanticSHA: semanticSHA, OriginalPlanSHA: prepared.PlanSHA256(), RankedPlanSHA: prepared.PlanSHA256(),
		ArtifactSHA: model.artifactSHA, ModelFingerprint: "sha256:" + model.fingerprint(), FeatureVersion: model.featureVersion(),
		Inputs: []PathContextInput{}, Scope: "immutable source and authored intent; initially empty observation channels; input_sha256 hashes the versioned little-endian FP32 array; no predictions or candidate evaluation"}
	if model.execution != nil {
		r.Schema = "gooo/compiler-execution-model-context/v1"
	}
	if model.flow != nil {
		r.Schema = "gooo/compiler-flow-model-context/v1"
	}
	input, err := model.initialInput(document, prepared)
	if err != nil {
		return prepared, r, false, err
	}
	for _, choice := range document.Plan.Decisions {
		if err := ctx.Err(); err != nil {
			return prepared, r, false, err
		}
		exported, err := model.exportInput(prepared, input, choice)
		if err != nil {
			r.Status, r.Reason, r.DeclinedDecision = "DECLINED_TO_DETERMINISTIC", err.Error(), choice.ID
			return prepared, r, true, nil
		}
		r.Inputs = append(r.Inputs, exported.PathContextInput)
	}
	return prepared, r, false, ctx.Err()
}

func exportedConditionInputs(document pathplan.Document, prepared *pathplan.PreparedPlan, model *conditionPathModel) ([]ExportedPathInput, error) {
	if model.contract != nil {
		return exportedContractInputs(document, prepared, model.orderedContract())
	}
	input, err := model.initialInput(document, prepared)
	if err != nil {
		return nil, err
	}
	result := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := model.exportInput(prepared, input, choice)
		if err != nil {
			return nil, err
		}
		result = append(result, exported)
	}
	return result, nil
}
