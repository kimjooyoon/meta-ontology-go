package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func (m *conditionPathModel) sourceRelativeV3Contract() bool {
	return m.contract != nil && m.contract.ArtifactSchema() == contractdecision.SourceRelativeV3Schema
}

func (m *conditionPathModel) sourceRelativeV3ProseNeutralContract() bool {
	return m.contract != nil && m.contract.ArtifactSchema() == contractdecision.SourceRelativeV3ProseNeutralSchema
}

func exportedSourceRelativeV3ContractInput(input *pathplan.ContractInput, choice pathplan.Choice) (ExportedPathInput, error) {
	return exportedSourceRelativeV3Input(input, choice, false)
}

func exportedSourceRelativeV3Input(input *pathplan.ContractInput, choice pathplan.Choice, proseNeutral bool) (ExportedPathInput, error) {
	intent := digest([]byte(choice.Intent))
	r := ExportedPathInput{DecisionID: choice.ID, OriginalIntentSHA: intent, NaturalIntentSHA: intent, Text: choice.Intent}
	var features [528]float32
	encode := input.SourceRelativeV3FeaturesInto
	if proseNeutral {
		encode = input.SourceRelativeV3ProseNeutralFeaturesInto
	}
	if err := encode(choice.ID, &features); err != nil {
		return r, err
	}
	r.SourceRelativeFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return r, nil
}

func exportedSourceRelativeV3ContractInputs(document pathplan.Document, prepared *pathplan.PreparedPlan) ([]ExportedPathInput, error) {
	return exportedSourceRelativeV3Inputs(document, prepared, false)
}

func exportedSourceRelativeV3Inputs(document pathplan.Document, prepared *pathplan.PreparedPlan, proseNeutral bool) ([]ExportedPathInput, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	rows := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := exportedSourceRelativeV3Input(input, choice, proseNeutral)
		if err != nil {
			return nil, err
		}
		branch, err := prepared.CanonicalOrderedBranchContextV2(choice.ID)
		if err != nil {
			return nil, err
		}
		conditions, err := input.SourceRelativeV2Conditions(choice.ID)
		if err != nil {
			return nil, err
		}
		outputs, err := input.SourceRelativeOutputs(choice.ID)
		if err != nil {
			return nil, err
		}
		if !branch.Available || !conditions.Available || !outputs.Available {
			return nil, fmt.Errorf("V3 source explanation unavailable: %s; %s; %s", branch.Reason, conditions.Reason, outputs.Reason)
		}
		exported.CanonicalBranch, exported.SourceRelativeConditions, exported.SourceRelativeOutputs = &branch, &conditions, &outputs
		rows = append(rows, exported)
	}
	return rows, nil
}
