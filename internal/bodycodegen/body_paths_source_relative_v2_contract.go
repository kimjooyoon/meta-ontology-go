package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func (m *conditionPathModel) sourceRelativeV2Contract() bool {
	return m.contract != nil && m.contract.ArtifactSchema() == contractdecision.SourceRelativeV2Schema
}

func exportedSourceRelativeV2ContractInput(input *pathplan.ContractInput, choice pathplan.Choice) (ExportedPathInput, error) {
	intent := digest([]byte(choice.Intent))
	r := ExportedPathInput{DecisionID: choice.ID,
		OriginalIntentSHA: intent, NaturalIntentSHA: intent, Text: choice.Intent}
	var features [528]float32
	if err := input.SourceRelativeV2FeaturesInto(choice.ID, &features); err != nil {
		return r, err
	}
	r.SourceRelativeFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return r, nil
}

func exportedSourceRelativeV2ContractInputs(document pathplan.Document, prepared *pathplan.PreparedPlan) ([]ExportedPathInput, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	r := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := exportedSourceRelativeV2ContractInput(input, choice)
		if err != nil {
			return nil, err
		}
		branch, err := prepared.CanonicalOrderedBranchContextV2(choice.ID)
		if err != nil {
			return nil, err
		}
		relations, err := input.SourceRelativeV2Conditions(choice.ID)
		if err != nil {
			return nil, err
		}
		if !branch.Available || !relations.Available {
			return nil, fmt.Errorf("V2 source explanation unavailable: %s; %s", branch.Reason, relations.Reason)
		}
		exported.CanonicalBranch, exported.SourceRelativeConditions = &branch, &relations
		r = append(r, exported)
	}
	return r, nil
}
