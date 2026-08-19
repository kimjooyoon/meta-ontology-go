package pressureindependence

type validation struct {
	decision Decision
	reason   Reason
	records  map[string]PressureRecord
}

func Evaluate(input Input) Output {
	output := Output{Schema: SchemaV1, FixtureID: input.FixtureID, InputDigest: CanonicalInputDigest(input)}
	check := validate(input)
	if check.decision != "" {
		return finishUnknown(output, input, check.decision, check.reason)
	}
	groups := make(map[string][]string)
	for _, id := range input.RequiredPressureIDs {
		groups[check.records[id].IndependenceGroupID] = append(groups[check.records[id].IndependenceGroupID], id)
	}
	output.DistinctGroupCount = uint64(len(groups))
	if uint64(len(groups)) < input.MinimumIndependent ||
		uint64(len(groups)) < effectiveK(input.RequestedK) {
		return finishUnknown(output, input, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	selected := selectRepresentatives(groups, input.RequestedK)
	output.SelectedIDs = selected
	output.UnselectedIDs = subtract(input.RequiredPressureIDs, selected)
	output.CostReceipt = receipt(input, len(selected))
	if !withinCeilings(output.CostReceipt, input.ResourceCeilings) {
		return finishUnknown(output, input, DecisionUnknown, ReasonInvalidResourceReceipt)
	}
	output.Decision = DecisionPass
	output.Reason = ReasonNone
	output.FullSuiteRequired = false
	output.ProofValid = finiteProof(input, output)
	return seal(output)
}
