package semanticdeltareceipt

import "sort"

func claimDelta(before, after projectedSource) ClaimDelta {
	result := ClaimDelta{Status: "KNOWN"}
	left, right := claimsBySlot(before.claims), claimsBySlot(after.claims)
	for slot, beforeClaims := range left {
		afterClaims := append([]Claim(nil), right[slot]...)
		sort.Slice(beforeClaims, func(i, j int) bool { return beforeClaims[i].ID < beforeClaims[j].ID })
		sort.Slice(afterClaims, func(i, j int) bool { return afterClaims[i].ID < afterClaims[j].ID })
		pairs := min(len(afterClaims), len(beforeClaims))
		for index := range pairs {
			if !claimMeaningEqual(beforeClaims[index], afterClaims[index]) {
				result.Changed = append(result.Changed, ClaimChange{ID: preservationClaimID(beforeClaims[index]), Before: beforeClaims[index], After: afterClaims[index]})
			}
		}
		result.Removed = append(result.Removed, beforeClaims[pairs:]...)
		result.Added = append(result.Added, afterClaims[pairs:]...)
	}
	for slot, afterClaims := range right {
		if _, ok := left[slot]; !ok {
			result.Added = append(result.Added, afterClaims...)
		}
	}
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].ID < result.Added[j].ID })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].ID < result.Removed[j].ID })
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].ID < result.Changed[j].ID })
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
	return left.ID == right.ID && left.Kind == right.Kind && left.Subject == right.Subject && left.Predicate == right.Predicate && left.Object == right.Object && left.Status == right.Status && left.PropositionDigest == right.PropositionDigest
}

func claimsBySlot(values []Claim) map[string][]Claim {
	result := make(map[string][]Claim)
	for _, value := range values {
		result[value.Subject+"\x00"+value.Predicate] = append(result[value.Subject+"\x00"+value.Predicate], value)
	}
	return result
}

func hasSemanticDelta(structural StructuralDelta, claims ClaimDelta) bool {
	return len(structural.AddedNodes)+len(structural.RemovedNodes)+len(structural.AddedFacts)+len(structural.RemovedFacts)+len(claims.Added)+len(claims.Removed)+len(claims.Changed) > 0
}
