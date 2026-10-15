package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func (m *conditionPathModel) sourceRelativeContract() bool {
	return m.contract != nil && m.contract.ArtifactSchema() == contractdecision.SourceRelativeRequirementSchema
}

func exportedSourceRelativeContractInput(input *pathplan.ContractInput, choice pathplan.Choice) (ExportedPathInput, error) {
	intent := digest([]byte(choice.Intent))
	r := ExportedPathInput{DecisionID: choice.ID,
		OriginalIntentSHA: intent, NaturalIntentSHA: intent, Text: choice.Intent}
	var features [contractdecision.OrderedFeatureDim]float32
	if err := input.SourceRelativeOrderedSourceFeaturesInto(choice.ID, &features); err != nil {
		return r, err
	}
	r.SourceRelativeFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return r, nil
}

// Explicit exports include the source explanation; normal receipts keep only
// hashes. The SDK derives relations from the same bound source and conditions
// used by the construction session, before prediction or candidate execution.
func exportedSourceRelativeContractInputs(document pathplan.Document, prepared *pathplan.PreparedPlan) ([]ExportedPathInput, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	r := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := exportedSourceRelativeContractInput(input, choice)
		if err != nil {
			return nil, err
		}
		branch, err := prepared.CanonicalOrderedBranchContext(choice.ID)
		if err != nil {
			return nil, err
		}
		relations, err := input.SourceRelativeConditions(choice.ID)
		if err != nil {
			return nil, err
		}
		if !branch.Available || !relations.Available {
			return nil, fmt.Errorf("source-relative explanation unavailable: %s; %s", branch.Reason, relations.Reason)
		}
		exported.CanonicalBranch, exported.SourceRelativeConditions = &branch, &relations
		r = append(r, exported)
	}
	return r, nil
}
