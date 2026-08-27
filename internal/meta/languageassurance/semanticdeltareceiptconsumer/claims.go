package semanticdeltareceiptconsumer

import "sort"

func claimDelta(before, after projectedSource) ClaimDelta {
	left, right := claimMap(before.claims), claimMap(after.claims)
	result := ClaimDelta{Status: "KNOWN"}
	for id, claim := range left {
		other, ok := right[id]
		if !ok {
			result.Removed = append(result.Removed, claim)
		} else if claimMeaningEqual(claim, other) == false {
			result.Changed = append(result.Changed, ClaimChange{ID: id, Before: claim, After: other})
		}
	}
	for id, claim := range right {
		if _, ok := left[id]; !ok {
			result.Added = append(result.Added, claim)
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
	return left.ID == right.ID && left.Subject == right.Subject && left.Predicate == right.Predicate && left.Object == right.Object && left.Status == right.Status
}

func hasSemanticDelta(structural StructuralDelta, claims ClaimDelta) bool {
	return len(structural.AddedNodes)+len(structural.RemovedNodes)+len(structural.AddedFacts)+len(structural.RemovedFacts)+len(claims.Added)+len(claims.Removed)+len(claims.Changed) > 0
}
