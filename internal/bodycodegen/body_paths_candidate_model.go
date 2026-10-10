package bodycodegen

import (
	"context"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/executiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Shape and version come from the decoded artifact. A v1 artifact never gains
// branch or output channels just because a newer runtime is available.
func (m *conditionPathModel) featureVersion() string {
	if m.orderedContract() {
		return contractdecision.OrderedSourceFeatureVersion
	}
	if m.contract != nil {
		return decision.RelationalFlowFeatureVersion
	}
	if m.flow != nil {
		return m.flow.FeatureVersion()
	}
	if m.execution != nil {
		return m.execution.FeatureVersion()
	}
	return m.model.FeatureVersion()
}

func (m *conditionPathModel) fingerprint() string {
	if m.contract != nil {
		return m.contract.Fingerprint()
	}
	if m.flow != nil {
		return m.flow.Fingerprint()
	}
	if m.execution != nil {
		return m.execution.Fingerprint()
	}
	return m.model.Fingerprint()
}

func (m *conditionPathModel) describe(info *RetainedModelInfo) {
	info.ArtifactSHA256, info.ModelFingerprint = m.artifactSHA, "sha256:"+m.fingerprint()
	info.FeatureVersion = m.featureVersion()
	info.ModelSchema, info.ResidentTensorBytes = conditionModelSchema, conditionModelTensorBytes
	if m.contract != nil {
		info.ModelSchema, info.ResidentTensorBytes = m.contract.ArtifactSchema(), contractdecision.ParameterCount*4
	}
	switch m.contract.(type) {
	case *contractdecision.ChoiceModel:
		info.ResidentTensorBytes = contractdecision.ChoiceParameterCount * 4
	case *contractdecision.OrderedRequirementModel:
		info.ResidentTensorBytes = contractdecision.OrderedRequirementParameterCount * 4
	case *contractdecision.InteractionRequirementModel:
		info.ResidentTensorBytes = contractdecision.InteractionRequirementParameterCount * 4
	}
	if m.execution != nil {
		info.ModelSchema, info.ResidentTensorBytes = executiondecision.Schema, executiondecision.ParameterCount*4
	}
	if m.flow != nil {
		info.ModelSchema, info.ResidentTensorBytes = m.flow.ArtifactSchema(), flowdecision.ParameterCount*4
	}
}

func decodeCandidateModel(raw []byte, schema string) (typedPathModel, error) {
	m := &conditionPathModel{artifactSHA: digest(raw)}
	var err error
	if schema == contractdecision.InteractionRequirementSchema {
		m.contract, err = contractdecision.DecodeInteractionRequirementConditioned(raw)
	} else if schema == contractdecision.OrderedRequirementSchema {
		m.contract, err = contractdecision.DecodeOrderedRequirementConditioned(raw)
	} else if schema == contractdecision.ChoiceSchema {
		m.contract, err = contractdecision.DecodeChoiceConditioned(raw)
	} else if schema == contractdecision.Schema || schema == contractdecision.PoolingSchema {
		m.contract, err = contractdecision.Decode(raw)
	} else if schema == flowdecision.Schema || schema == flowdecision.ActivationSchema {
		m.flow, err = flowdecision.Decode(raw)
	} else if schema == executiondecision.Schema {
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
	if m.execution != nil || m.flow != nil {
		return prepared.InitialExecutionInput(document.TestCases)
	}
	return prepared.InitialConditionInput()
}

func (m *conditionPathModel) search(ctx context.Context, prepared *pathplan.PreparedPlan, cases []pathplan.TestCase,
	total, step int, seed string, rounds int) (pathplan.SearchResult, *bodyplan.Program, []pathplan.ConditionProgress, []pathplan.ConditionRanking, error) {
	if m.flow != nil {
		return prepared.SearchFlowBatches(ctx, m.flow, cases, total, step, seed, rounds)
	}
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
	if m.flow != nil {
		return m.exportFlowInput(prepared, input, choice.ID, out)
	} else if m.execution != nil {
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

func (m *conditionPathModel) exportFlowInput(prepared *pathplan.PreparedPlan, input *pathplan.ConditionInput,
	id string, out ExportedPathInput) (ExportedPathInput, error) {
	var features [decision.ExecutionFlowFeatureDim]float32
	var err error
	if m.featureVersion() == decision.SemanticFlowFeatureVersion || m.featureVersion() == decision.RelationalFlowFeatureVersion {
		var semantic pathplan.SemanticBranchContext
		semantic, err = prepared.SemanticBranchContext(id)
		if err != nil {
			return out, err
		}
		out.SemanticFlow = &semantic
		out.SourceFeatureSHA = digest(semantic.Source[:])
		if m.featureVersion() == decision.RelationalFlowFeatureVersion {
			err = input.ExecutionRelationalFlowFeaturesInto(id, &features)
		} else {
			err = input.ExecutionSemanticFlowFeaturesInto(id, &features)
		}
	} else {
		err = input.ExecutionFlowFeaturesInto(id, &features)
	}
	if err != nil {
		return out, err
	}
	out.FlowFeatures, out.Bytes, out.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return out, nil
}
