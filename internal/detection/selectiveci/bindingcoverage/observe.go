package bindingcoverage

// Observe evaluates explicit binding partitions without side effects or
// inferred coverage. Classify is a descriptive alias for the same observer.
func Observe(input Input) Output {
	input = normalizeInput(input)
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return seal(baseOutput(input, 0, ""), DecisionUnknown, ReasonEvaluatorError)
	}
	result := baseOutput(input, uint64(len(canonical)), digestBytes(canonical))
	if input.RequiredBindings == nil || input.Partitions == nil || input.PrecedenceRegistry == nil {
		return seal(result, DecisionUnknown, ReasonMissingInput)
	}
	// Shape counts describe the canonical input itself, not validated semantic
	// success. Populate them before any schema, header, registry, or reference
	// validation so UNKNOWN cannot launder partial validation counts.
	if !populateShapeCounts(&result, len(input.RequiredBindings), len(input.Partitions)) {
		return seal(result, DecisionUnknown, ReasonWorkOverflow)
	}
	if input.SchemaVersion != SchemaVersion {
		return seal(result, DecisionUnknown, ReasonUnknownSchema)
	}
	if reason := validateHeader(input); reason != "" {
		return seal(result, DecisionUnknown, reason)
	}

	precedencePairs, reason := validatePrecedence(input.PrecedenceRegistry)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	bindingPairs, reason := validateBindings(input.RequiredBindings, precedencePairs)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	match, mismatch, reason := validatePartitions(input.Partitions, bindingPairs)
	if reason != "" {
		return seal(result, DecisionUnknown, reason)
	}
	result.MissingMatchBindingIDs, result.MissingMismatchBindingIDs = missingBindings(input.RequiredBindings, match, mismatch)
	result.MatchCoveredCount = result.RequiredBindingCount - uint64(len(result.MissingMatchBindingIDs))
	result.MismatchCoveredCount = result.RequiredBindingCount - uint64(len(result.MissingMismatchBindingIDs))
	if result.RequiredBindingCount == 0 {
		return seal(result, DecisionIncomplete, ReasonZeroDenominator)
	}
	hasMissingMatch := len(result.MissingMatchBindingIDs) != 0
	hasMissingMismatch := len(result.MissingMismatchBindingIDs) != 0
	switch {
	case hasMissingMatch && hasMissingMismatch:
		return seal(result, DecisionIncomplete, ReasonMissingMatchAndMismatch)
	case hasMissingMatch:
		return seal(result, DecisionIncomplete, ReasonMissingMatch)
	case hasMissingMismatch:
		return seal(result, DecisionIncomplete, ReasonMissingMismatch)
	}
	return seal(result, DecisionExact, ReasonComplete)
}

func Classify(input Input) Output { return Observe(input) }

func Evaluate(input Input) Output { return Observe(input) }

func baseOutput(input Input, inputBytes uint64, inputDigest string) Output {
	return Output{SchemaVersion: input.SchemaVersion, ContractID: input.ContractID, SnapshotDigest: input.SnapshotDigest, ExpectedSnapshotDigest: input.ExpectedSnapshotDigest, InputDigest: inputDigest, InputBytes: inputBytes}
}

func seal(output Output, decision Decision, reason Reason) Output {
	output = normalizeOutput(output)
	output.Decision = decision
	output.Reason = reason
	output.CanonicalDigest = output.StableDigest()
	return output
}
