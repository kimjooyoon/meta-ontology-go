package fullsoundness

func commandMap(values []Command, obligations map[string]ObligationAuthority) (map[string]Command, Reason) {
	result := make(map[string]Command, len(values))
	for _, value := range values {
		if !validID(value.ID) || value.ObligationIDs == nil {
			return nil, ReasonFullSuiteRequired
		}
		if _, exists := result[value.ID]; exists {
			return nil, ReasonFullSuiteRequired
		}
		ids, unique := stringSet(value.ObligationIDs)
		if !unique {
			return nil, ReasonFullSuiteRequired
		}
		for id := range ids {
			if !validID(id) {
				return nil, ReasonFullSuiteRequired
			}
			if _, registered := obligations[id]; !registered {
				return nil, ReasonUnregisteredObligation
			}
		}
		result[value.ID] = value
	}
	return result, ""
}
func impactedSet(values []string, obligations map[string]ObligationAuthority) (map[string]struct{}, Reason) {
	result, unique := stringSet(values)
	if !unique {
		return nil, ReasonFullSuiteRequired
	}
	for id := range result {
		authority, registered := obligations[id]
		if !validID(id) {
			return nil, ReasonFullSuiteRequired
		}
		if !registered {
			return nil, ReasonUnregisteredObligation
		}
		if authority != AuthorityAuthoritative {
			return nil, ReasonUnprovableObligation
		}
	}
	return result, ""
}
func knownSet(values []string, commands map[string]Command) (map[string]struct{}, Reason) {
	result, unique := stringSet(values)
	if !unique {
		return nil, ReasonFullSuiteRequired
	}
	for id := range result {
		if !validID(id) || commands[id].ID == "" {
			return nil, ReasonFullSuiteRequired
		}
	}
	return result, ""
}
func validateReceipt(input Input, state evaluationState) Reason {
	receipt := input.SelectionReceipt
	if !validDigest(receipt.SnapshotDigest) || !validDigest(receipt.PolicyDigest) || !validDigest(receipt.RegistryDigest) || !validDigest(receipt.SelectionDigest) {
		return ReasonFullSuiteRequired
	}
	if receipt.SnapshotDigest != input.SnapshotDigest || receipt.PolicyDigest != input.PolicyDigest || receipt.RegistryDigest != input.RegistryDigest || receipt.SelectionDigest != input.SelectionDigest {
		return ReasonDigestBindingMismatch
	}
	_, reason := knownSet(receipt.CommandIDs, state.commands)
	return reason
}
