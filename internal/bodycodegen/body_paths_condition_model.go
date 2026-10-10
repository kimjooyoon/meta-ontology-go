package bodycodegen

import (
	"context"
	"encoding/binary"
	"math"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const conditionModelSchema = conditiondecision.Schema
const conditionModelTensorBytes = conditiondecision.ParameterCount * 4

// The raw artifact identity is distinct from its formatting-independent model
// fingerprint. Both bind the same immutable, once-decoded weight arrays.
type conditionPathModel struct {
	model       *conditiondecision.Model
	artifactSHA string
}

func conditionFeatureDigest(features [decision.FeatureDim]float32) string {
	var raw [decision.FeatureDim * 4]byte
	for i, value := range features {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(value))
	}
	return digest(raw[:])
}

func prepareConditionModelContext(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	model *conditionPathModel, activityID, semanticSHA string) (*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	r := &PathModelContextReceipt{Schema: "gooo/compiler-condition-model-context/v1", Status: "ENCODED", ActivityID: activityID,
		SourceSemanticSHA: semanticSHA, OriginalPlanSHA: prepared.PlanSHA256(), RankedPlanSHA: prepared.PlanSHA256(),
		ArtifactSHA: model.artifactSHA, ModelFingerprint: "sha256:" + model.model.Fingerprint(), FeatureVersion: decision.ConditionChannelFeatureVersion,
		Inputs: []PathContextInput{}, Scope: "immutable source features and full authored intent in independent array regions; initially empty condition channel; input_sha256 hashes the 1024 little-endian FP32 feature bytes; no predictions or candidate evaluation"}
	input, err := prepared.InitialConditionInput()
	if err != nil {
		return prepared, r, false, err
	}
	for _, choice := range document.Plan.Decisions {
		if err := ctx.Err(); err != nil {
			return prepared, r, false, err
		}
		var features [decision.FeatureDim]float32
		if err := input.FeaturesInto(choice.ID, &features); err != nil {
			r.Status, r.Reason, r.DeclinedDecision = "DECLINED_TO_DETERMINISTIC", err.Error(), choice.ID
			return prepared, r, true, nil
		}
		source, err := prepared.SourceFeatures(choice.ID)
		if err != nil {
			return prepared, r, false, err
		}
		intentSHA := digest([]byte(choice.Intent))
		r.Inputs = append(r.Inputs, PathContextInput{DecisionID: choice.ID, OriginalIntentSHA: intentSHA, NaturalIntentSHA: intentSHA,
			InputSHA: conditionFeatureDigest(features), SourceFeatureSHA: digest(source[:]), Bytes: decision.FeatureDim * 4})
	}
	return prepared, r, false, ctx.Err()
}

func exportedConditionInputs(document pathplan.Document, prepared *pathplan.PreparedPlan) ([]ExportedPathInput, error) {
	input, err := prepared.InitialConditionInput()
	if err != nil {
		return nil, err
	}
	result := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		var features [decision.FeatureDim]float32
		if err := input.FeaturesInto(choice.ID, &features); err != nil {
			return nil, err
		}
		source, err := prepared.SourceFeatures(choice.ID)
		if err != nil {
			return nil, err
		}
		intentSHA := digest([]byte(choice.Intent))
		result = append(result, ExportedPathInput{DecisionID: choice.ID, OriginalIntentSHA: intentSHA,
			NaturalIntentSHA: intentSHA, InputSHA: conditionFeatureDigest(features), SourceFeatureSHA: digest(source[:]), Bytes: decision.FeatureDim * 4,
			Text: choice.Intent, Features: &features})
	}
	return result, nil
}
