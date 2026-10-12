package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func (m *conditionPathModel) sourceRelativeV4Contract() bool {
	return m.contract != nil && m.contract.ArtifactSchema() == contractdecision.SourceRelativeV4Schema
}

func exportedSourceRelativeV4Input(input *pathplan.ContractInput, choice pathplan.Choice) (ExportedPathInput, error) {
	intent := digest([]byte(choice.Intent))
	r := ExportedPathInput{DecisionID: choice.ID, OriginalIntentSHA: intent, NaturalIntentSHA: intent, Text: choice.Intent}
	var features [528]float32
	if err := input.SourceRelativeV4FeaturesInto(choice.ID, &features); err != nil {
		return r, err
	}
	r.SourceRelativeFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return r, nil
}

func exportedSourceRelativeV4Inputs(document pathplan.Document, prepared *pathplan.PreparedPlan) ([]ExportedPathInput, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	rows := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := exportedSourceRelativeV4Input(input, choice)
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
		owned, err := input.BranchOutputs(choice.ID)
		if err != nil {
			return nil, err
		}
		if !branch.Available || !conditions.Available || !outputs.Available || !owned.Available {
			return nil, fmt.Errorf("V4 source explanation unavailable: %s; %s; %s; %s",
				branch.Reason, conditions.Reason, outputs.Reason, owned.Reason)
		}
		exported.CanonicalBranch, exported.SourceRelativeConditions, exported.SourceRelativeOutputs = &branch, &conditions, &outputs
		exported.BranchOutputs = &owned
		rows = append(rows, exported)
	}
	return rows, nil
}
