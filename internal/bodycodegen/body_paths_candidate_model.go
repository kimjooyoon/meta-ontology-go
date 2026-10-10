package bodycodegen

import (
	"context"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/executiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Shape and version come from the decoded artifact. A v1 artifact never gains
// branch or output channels just because a newer runtime is available.
func (m *conditionPathModel) featureVersion() string {
	if m.execution != nil {
		return m.execution.FeatureVersion()
	}
	return m.model.FeatureVersion()
}

func (m *conditionPathModel) fingerprint() string {
	if m.execution != nil {
		return m.execution.Fingerprint()
	}
	return m.model.Fingerprint()
}

func (m *conditionPathModel) describe(info *RetainedModelInfo) {
	info.ArtifactSHA256, info.ModelFingerprint = m.artifactSHA, "sha256:"+m.fingerprint()
	info.FeatureVersion = m.featureVersion()
	info.ModelSchema, info.ResidentTensorBytes = conditionModelSchema, conditionModelTensorBytes
	if m.execution != nil {
		info.ModelSchema, info.ResidentTensorBytes = executiondecision.Schema, executiondecision.ParameterCount*4
	}
}

func decodeCandidateModel(raw []byte, schema string) (typedPathModel, error) {
	m := &conditionPathModel{artifactSHA: digest(raw)}
	var err error
	if schema == executiondecision.Schema {
		m.execution, err = executiondecision.Decode(raw)
	} else {
		m.model, err = conditiondecision.Decode(raw)
	}
	if err != nil {
		return typedPathModel{}, err
	}
	return typedPathModel{condition: m}, nil
}

func (m *conditionPathModel) initialInput(document pathplan.Document, prepared *pathplan.PreparedPlan) (*pathplan.ConditionInput, error) {
	if m.execution != nil {
		return prepared.InitialExecutionInput(document.TestCases)
	}
	return prepared.InitialConditionInput()
}

func (m *conditionPathModel) search(ctx context.Context, prepared *pathplan.PreparedPlan, cases []pathplan.TestCase,
	total, step int, seed string, rounds int) (pathplan.SearchResult, *bodyplan.Program, []pathplan.ConditionProgress, []pathplan.ConditionRanking, error) {
	if m.execution != nil {
		return prepared.SearchExecutionBatches(ctx, m.execution, cases, total, step, seed, rounds)
	}
	return prepared.SearchConditionBatches(ctx, m.model, cases, total, step, seed, rounds)
}

func (m *conditionPathModel) exportInput(prepared *pathplan.PreparedPlan, input *pathplan.ConditionInput,
	choice pathplan.Choice) (ExportedPathInput, error) {
	intentSHA := digest([]byte(choice.Intent))
	out := ExportedPathInput{DecisionID: choice.ID,
		OriginalIntentSHA: intentSHA, NaturalIntentSHA: intentSHA, Text: choice.Intent}
	source, err := prepared.SourceFeatures(choice.ID)
	if err != nil {
		return out, err
	}
	out.SourceFeatureSHA = digest(source[:])
	if m.execution != nil {
		var features [decision.ExecutionFeatureDim]float32
		if err := input.ExecutionFeaturesInto(choice.ID, &features); err != nil {
			return out, err
		}
		out.ExecutionFeatures, out.Bytes, out.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	} else {
		var features [decision.FeatureDim]float32
		if err := input.FeaturesIntoVersion(choice.ID, m.featureVersion(), &features); err != nil {
			return out, err
		}
		out.Features, out.Bytes, out.InputSHA = &features, len(features)*4, conditionFeatureDigest(features)
	}
	return out, nil
}
