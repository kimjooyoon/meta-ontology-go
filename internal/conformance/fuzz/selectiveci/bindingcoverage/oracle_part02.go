package bindingcoverage

func Evaluate(input Input) Result {
	vector, overflow := baseVector(input)
	if overflow {
		return finish(vector, DecisionUnknown, "COUNT_OVERFLOW")
	}
	if reason := validateInput(input); reason != "" {
		return finish(vector, DecisionUnknown, reason)
	}
	if len(input.RequiredBindings) == 0 {
		return finish(vector, DecisionIncomplete, "ZERO_DENOMINATOR")
	}
	missingMatch, missingMismatch := missingEvidence(input)
	vector.MissingMatch = missingMatch
	vector.MissingMismatch = missingMismatch
	if len(missingMatch) != 0 || len(missingMismatch) != 0 {
		return finish(vector, DecisionIncomplete, incompleteReason(missingMatch, missingMismatch))
	}
	return finish(vector, DecisionExact, "COMPLETE")
}
func baseVector(input Input) (Vector, bool) {
	endpointCount, overflow := endpointReferenceCount(input.RequiredBindings)
	work, workOverflow := safeAdd(int64(len(input.RequiredBindings)), int64(len(input.Partitions)))
	overflow = overflow || workOverflow
	if !overflow {
		work, overflow = safeAdd(work, int64(endpointCount))
	}
	canonical, marshalErr := canonicalInputJSON(input)
	if marshalErr != nil {
		overflow = true
	}
	return Vector{
		RequiredBindingCount:   int64(len(input.RequiredBindings)),
		PartitionCount:         int64(len(input.Partitions)),
		EndpointReferenceCount: endpointCount,
		InputBytes:             int64(len(canonical)),
		WorkUnits:              work,
		MissingMatch:           []string{},
		MissingMismatch:        []string{},
		ExecutionAuthorized:    false,
		CIAuthorized:           false,
	}, overflow
}
func validateInput(input Input) string {
	if input.Schema != SchemaV1 {
		return "UNKNOWN_SCHEMA"
	}
	if !validDigest(input.SnapshotDigest) || input.SnapshotDigest != input.ExpectedDigest {
		return "STALE_OR_BAD_DIGEST"
	}
	precedence, reason := validatePrecedence(input.Precedence)
	if reason != "" {
		return reason
	}
	bindings, reason := validateBindings(input.RequiredBindings, precedence)
	if reason != "" {
		return reason
	}
	return validatePartitions(input.Partitions, bindings, precedence)
}
