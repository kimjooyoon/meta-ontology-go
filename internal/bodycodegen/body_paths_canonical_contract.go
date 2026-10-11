package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func (m *conditionPathModel) canonicalContract() bool {
	return m.contract != nil && m.contract.ArtifactSchema() == contractdecision.CanonicalInteractionRequirementSchema
}

func exportedCanonicalContractInput(input *pathplan.ContractInput, choice pathplan.Choice) (ExportedPathInput, error) {
	intent := digest([]byte(choice.Intent))
	r := ExportedPathInput{DecisionID: choice.ID,
		OriginalIntentSHA: intent, NaturalIntentSHA: intent, Text: choice.Intent}
	var features [contractdecision.OrderedFeatureDim]float32
	if err := input.CanonicalOrderedSourceFeaturesInto(choice.ID, &features); err != nil {
		return r, err
	}
	r.CanonicalFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return r, nil
}

func exportedCanonicalContractInputs(document pathplan.Document, prepared *pathplan.PreparedPlan) ([]ExportedPathInput, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	r := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := exportedCanonicalContractInput(input, choice)
		if err != nil {
			return nil, err
		}
		view, err := prepared.CanonicalOrderedBranchContext(choice.ID)
		if err != nil {
			return nil, err
		}
		if !view.Available {
			return nil, fmt.Errorf("canonical explanation unavailable: %s", view.Reason)
		}
		exported.CanonicalBranch = &view
		r = append(r, exported)
	}
	return r, nil
}
