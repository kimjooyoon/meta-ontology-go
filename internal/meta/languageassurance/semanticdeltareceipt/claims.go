package semanticdeltareceipt

import "sort"

func claimDelta(before, after projectedSource) ClaimDelta {
	result := ClaimDelta{Status: "KNOWN"}
	left, right := claimsBySlot(before.claims), claimsBySlot(after.claims)
	slots := make(map[string]bool, len(left)+len(right))
	for slot := range left {
		slots[slot] = true
	}
	for slot := range right {
		slots[slot] = true
	}
	ordered := make([]string, 0, len(slots))
	for slot := range slots {
		ordered = append(ordered, slot)
	}
	sort.Strings(ordered)
	for _, slot := range ordered {
		beforeClaims, afterClaims := left[slot], right[slot]
		if len(beforeClaims) > 1 || len(afterClaims) > 1 {
			result.Status, result.Reason = "UNKNOWN", ReasonAmbiguous
			result.Ambiguous = append(result.Ambiguous, ClaimMatch{Slot: slot, BeforeCount: len(beforeClaims), AfterCount: len(afterClaims), Reason: ReasonAmbiguous})
			continue
		}
		if len(beforeClaims) == 1 && len(afterClaims) == 1 {
			if !claimMeaningEqual(beforeClaims[0], afterClaims[0]) {
				result.Changed = append(result.Changed, ClaimChange{ID: preservationClaimID(beforeClaims[0]), Before: beforeClaims[0], After: afterClaims[0]})
			}
			continue
		}
		result.Removed = append(result.Removed, beforeClaims...)
		result.Added = append(result.Added, afterClaims...)
	}
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].ID < result.Added[j].ID })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].ID < result.Removed[j].ID })
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].ID < result.Changed[j].ID })
	sort.Slice(result.Ambiguous, func(i, j int) bool { return result.Ambiguous[i].Slot < result.Ambiguous[j].Slot })
	return result
}

func claimMap(values []Claim) map[string]Claim {
	result := make(map[string]Claim, len(values))
	for _, value := range values {
		result[value.ID] = value
	}
	return result
}

func claimMeaningEqual(left, right Claim) bool {
	return left.ClaimTypeID == right.ClaimTypeID && left.Kind == right.Kind && left.Subject == right.Subject && left.Predicate == right.Predicate && left.Object == right.Object && left.NormalizedProposition == right.NormalizedProposition
}

func claimsBySlot(values []Claim) map[string][]Claim {
	result := make(map[string][]Claim)
	for _, value := range values {
		result[value.Subject+"\x00"+value.Predicate] = append(result[value.Subject+"\x00"+value.Predicate], value)
	}
	for slot := range result {
		sort.Slice(result[slot], func(i, j int) bool { return result[slot][i].ID < result[slot][j].ID })
	}
	return result
}

func semanticComponentDelta(before, after projectedSource) SemanticComponentDelta {
	result := SemanticComponentDelta{Status: "KNOWN"}
	left, right := componentMap(before.semanticComponents), componentMap(after.semanticComponents)
	for key, value := range right {
		old, ok := left[key]
		if !ok {
			result.Added = append(result.Added, value)
			continue
		}
		if old.Object != value.Object || old.PropositionDigest != value.PropositionDigest {
			result.Changed = append(result.Changed, ComponentChange{ID: value.ID, Before: old, After: value})
		}
	}
	for key, value := range left {
		if _, ok := right[key]; !ok {
			result.Removed = append(result.Removed, value)
		}
	}
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].ID < result.Added[j].ID })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].ID < result.Removed[j].ID })
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].ID < result.Changed[j].ID })
	return result
}

func componentMap(values []SemanticComponent) map[string]SemanticComponent {
	result := make(map[string]SemanticComponent, len(values))
	for _, value := range values {
		result[value.Kind+"\x00"+value.Subject+"\x00"+value.Predicate] = value
	}
	return result
}

func hasSemanticDelta(structural StructuralDelta, components SemanticComponentDelta, claims ClaimDelta) bool {
	return len(structural.AddedNodes)+len(structural.RemovedNodes)+len(structural.AddedFacts)+len(structural.RemovedFacts)+len(components.Added)+len(components.Removed)+len(components.Changed)+len(claims.Added)+len(claims.Removed)+len(claims.Changed) > 0
}
