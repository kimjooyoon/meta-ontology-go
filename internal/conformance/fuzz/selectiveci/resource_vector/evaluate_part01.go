package resourcevector

type validationFailure struct {
	decision Decision
	reason   Reason
}
type indexedRecords struct {
	commands map[string]CommandRecord
	paths    map[string]PathRecord
	byCmd    map[string][]PathRecord
}
type vectorMeta struct {
	pathCount int
	allFinite bool
}

// Evaluate strictly replays canonical command and path records. It never
// trusts an expected vector and never imports a production selector.
func Evaluate(input Input) Output {
	output := Output{
		Schema: input.Schema, FixtureID: input.FixtureID, InputDigest: CanonicalInputDigest(input),
		LimitFailures: []string{}, FullSuiteRequired: true,
	}
	if failure := validate(input); failure.reason != "" {
		return finish(output, failure.decision, failure.reason)
	}
	indexed := index(input)
	selected, selectedMeta, ok := replayVector(indexed, input.SelectedCommandIDs)
	if !ok {
		return finish(output, DecisionUnknown, ReasonOverflow)
	}
	full, _, ok := replayVector(indexed, input.FullCommandIDs)
	if !ok {
		return finish(output, DecisionUnknown, ReasonOverflow)
	}
	output.Selected, output.Full = &selected, &full
	output.LimitFailures = append(output.LimitFailures, compareCeilings(selected, input.Ceilings.Selected, "selected")...)
	output.LimitFailures = append(output.LimitFailures, compareCeilings(full, input.Ceilings.Full, "full")...)
	if len(output.LimitFailures) != 0 {
		return finish(output, DecisionFailClosed, ReasonResourceLimitExceeded)
	}
	output.Decision, output.Reason, output.FullSuiteRequired = DecisionPass, ReasonNone, false
	output.ProofValid = selectedMeta.pathCount > 0 && selectedMeta.allFinite &&
		selected.FinitePROVPaths == uint64(selectedMeta.pathCount) &&
		selected.ClosureDenominator > 0 && selected.ClosureNumerator == selected.ClosureDenominator
	return finish(output, output.Decision, output.Reason)
}
func Replay(input Input) Output { return Evaluate(input) }
func finish(output Output, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	if decision != DecisionPass {
		output.FullSuiteRequired = true
		output.ProofValid = false
	}
	output.CanonicalOutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.CanonicalOutputDigest)
	return output
}
