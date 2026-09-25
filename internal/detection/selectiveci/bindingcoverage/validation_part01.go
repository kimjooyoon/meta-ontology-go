package bindingcoverage

func validateHeader(input Input) Reason {
	if input.ContractID == "" || input.SnapshotDigest == "" || input.ExpectedSnapshotDigest == "" {
		return ReasonMissingInput
	}
	if !validStableID(input.ContractID) {
		return ReasonInvalidID
	}
	if !validDigest(input.SnapshotDigest) || !validDigest(input.ExpectedSnapshotDigest) {
		return ReasonInvalidDigest
	}
	if input.SnapshotDigest != input.ExpectedSnapshotDigest {
		return ReasonSnapshotMismatch
	}
	return ""
}
func validatePrecedence(entries []PrecedenceEntry) (map[string]struct{}, Reason) {
	ranks := make(map[uint64]struct{}, len(entries))
	pairs := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if validateStageToken(entry.Stage) != "" || validateReasonToken(entry.Reason) != "" {
			return nil, ReasonInvalidPrecedence
		}
		if _, exists := ranks[entry.Rank]; exists {
			return nil, ReasonDuplicatePrecedence
		}
		pair := expectedPair(entry.Stage, entry.Reason)
		if _, exists := pairs[pair]; exists {
			return nil, ReasonDuplicatePrecedence
		}
		ranks[entry.Rank] = struct{}{}
		pairs[pair] = struct{}{}
	}
	return pairs, ""
}
func validateBindings(bindings []RequiredBinding, precedencePairs map[string]struct{}) (map[string]string, Reason) {
	ids := make(map[string]struct{}, len(bindings))
	bindingPairs := make(map[string]string, len(bindings))
	for _, binding := range bindings {
		if reason := validateID(binding.BindingID); reason != "" {
			return nil, reason
		}
		if _, exists := ids[binding.BindingID]; exists {
			return nil, ReasonDuplicateID
		}
		if reason := validateID(binding.FromFieldID); reason != "" {
			return nil, reason
		}
		if reason := validateID(binding.ToFieldID); reason != "" {
			return nil, reason
		}
		if !validKind(binding.Kind) {
			return nil, ReasonInvalidEnum
		}
		if reason := validateStageToken(binding.ExpectedStage); reason != "" {
			return nil, reason
		}
		if reason := validateReasonToken(binding.ExpectedReason); reason != "" {
			return nil, reason
		}
		if binding.FromFieldID == binding.ToFieldID {
			return nil, ReasonSelfLink
		}
		pair := expectedPair(binding.ExpectedStage, binding.ExpectedReason)
		if _, registered := precedencePairs[pair]; !registered {
			return nil, ReasonUnregisteredPair
		}
		ids[binding.BindingID] = struct{}{}
		bindingPairs[binding.BindingID] = pair
	}
	return bindingPairs, ""
}
