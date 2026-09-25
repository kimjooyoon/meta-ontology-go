package fullsoundness

type evaluationState struct {
	obligations      map[string]ObligationAuthority
	commands         map[string]Command
	impacted         map[string]struct{}
	selected         map[string]struct{}
	fullOutcomes     map[string]Outcome
	selectedOutcomes map[string]Outcome
	fullReceipts     map[string]ResourceReceipt
	selectedReceipts map[string]ResourceReceipt
}

// Evaluate is pure and leaves both authorization flags false in every result.
func Evaluate(input Input) Output {
	input = normalizeInput(input)
	result := baseOutput(input)
	populateCounts(&result, input)
	if missingRequiredInput(input) {
		return seal(result, DecisionUnknown, ReasonFullSuiteRequired)
	}
	state, reason := validateInput(input)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	if result.CommandCount == 0 {
		return seal(result, DecisionUnknown, ReasonZeroCommandDenominator)
	}
	if input.ExecutionAuthorized || input.CIAuthorized {
		return seal(result, DecisionUnsound, ReasonAuthorizationPresent)
	}
	if !sameStringSet(input.SelectedCommandIDs, input.SelectionReceipt.CommandIDs) {
		return seal(result, DecisionUnsound, ReasonSelectedSetMismatch)
	}
	if globalGuardOmitted(state) {
		return seal(result, DecisionUnsound, ReasonGlobalGuardOmitted)
	}
	result.SemanticEvaluated = true
	populateCommandLists(&result, state)
	if reason := checkSoundness(state); reason != "" {
		return seal(result, decisionFor(reason), reason)
	}
	vector, reason := resourceVector(state)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	result.ResourceVector = &vector
	return seal(result, DecisionSound, ReasonSound)
}
func Observe(input Input) Output  { return Evaluate(input) }
func Classify(input Input) Output { return Evaluate(input) }
func baseOutput(input Input) Output {
	return Output{SchemaVersion: input.SchemaVersion, SnapshotDigest: input.SnapshotDigest, PolicyDigest: input.PolicyDigest, RegistryDigest: input.RegistryDigest, SelectionDigest: input.SelectionDigest}
}
func seal(output Output, decision Decision, reason Reason) Output {
	output.Decision = decision
	output.Reason = reason
	output.ExecutionAuthorized = false
	output.CIAuthorized = false
	output = normalizeOutput(output)
	decisionDigest, err := output.DecisionStableDigest()
	if err != nil {
		return sealProjectionFailure(output)
	}
	output.DecisionDigest = decisionDigest
	output.CanonicalDigest = output.StableDigest()
	return output
}
