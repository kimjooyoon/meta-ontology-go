package improvement

func comparisonReason(before, after Snapshot, beforeState, afterState inspection) string {
	switch {
	case before.ContractSchema != after.ContractSchema:
		return "CONTRACT_SCHEMA_MISMATCH"
	case before.RegistryDigest != after.RegistryDigest:
		return "REGISTRY_DIGEST_MISMATCH"
	case before.Total != after.Total:
		return "DENOMINATOR_MISMATCH"
	}
	for id := range beforeState.statuses {
		if _, ok := afterState.statuses[id]; !ok {
			return "OBLIGATION_SET_MISMATCH"
		}
	}
	return ""
}

func quantify(result *Transition, before, after Snapshot, beforeState, afterState inspection) {
	result.Comparable = true
	result.CompletedDelta = after.Completed - before.Completed
	result.BasisPointsDelta = after.BasisPoints - before.BasisPoints
	result.BeforeUnresolved = beforeState.unresolved
	result.AfterUnresolved = afterState.unresolved
	for id, beforeStatus := range beforeState.statuses {
		afterStatus := afterState.statuses[id]
		if beforeStatus != Satisfied && afterStatus == Satisfied {
			result.Gains++
		}
		if beforeStatus == Satisfied && afterStatus != Satisfied {
			result.Regressions++
		}
	}
	resolved := result.BeforeUnresolved == 0 && result.AfterUnresolved == 0
	result.Indicators = indicators(*result)
	result.Proofs = proofs(true, true, resolved, result.Regressions == 0)
}
