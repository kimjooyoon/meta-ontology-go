package resourcevector

// EvaluateTypedConfigFullSuite is a fair baseline that runs every canonical
// command. It has its own aggregation loop and does not consume Output.
func EvaluateTypedConfigFullSuite(input Input) BaselineResult {
	return evaluateBaseline(input, input.FullCommandIDs, true, "typed-config+full-suite")
}

// EvaluatePlainDAGRetry is the plain impacted-DAG/retry baseline. For this
// fixture it sees the same selected IDs, but it remains a separate evaluator
// so an oracle cannot claim a benefit that the plain strategy already has.
func EvaluatePlainDAGRetry(input Input) BaselineResult {
	return evaluateBaseline(input, input.SelectedCommandIDs, false, "plain-dag/retry")
}
func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	typed := EvaluateTypedConfigFullSuite(input)
	plain := EvaluatePlainDAGRetry(input)
	comparison := Comparison{Oracle: oracle, TypedConfigFullSuite: typed, PlainDAGRetry: plain, Finding: NoUniqueBenefit}
	if oracle.Decision != DecisionPass || typed.Decision != DecisionPass || plain.Decision != DecisionPass ||
		oracle.Selected == nil || typed.Vector == nil || plain.Vector == nil {
		return comparison
	}
	if strictlyBetter(*oracle.Selected, typed.Vector) && strictlyBetter(*oracle.Selected, plain.Vector) {
		comparison.Finding = UniqueBenefitNotEstablished
	}
	return comparison
}
func evaluateBaseline(input Input, selectedIDs []string, fullSuite bool, name string) BaselineResult {
	result := BaselineResult{Name: name, Decision: DecisionUnknown, Reason: ReasonMissingInput, FullSuite: fullSuite}
	commands, ids, selected, failure := baselineSelection(input, selectedIDs)
	if failure.reason != "" {
		result.Decision, result.Reason = failure.decision, failure.reason
		return result
	}
	vector, resourceKnown, pressureKnown := baselineResources(commands, ids)
	affected, affectedFailure := baselineAffected(input, commands, ids)
	if affectedFailure.reason != "" {
		result.Vector = vector
		result.Decision, result.Reason = affectedFailure.decision, affectedFailure.reason
		return result
	}
	vector.AffectedStableIDs = new(affected)
	prov, provKnown := baselinePROV(input.Paths, selected)
	mergePartial(vector, prov)
	result.Vector = vector
	if !resourceKnown || !pressureKnown {
		result.Reason = ReasonMissingResource
		return result
	}
	if !provKnown {
		result.Reason = ReasonMissingPROV
		return result
	}
	result.Decision, result.Reason = DecisionPass, ReasonNone
	return result
}
