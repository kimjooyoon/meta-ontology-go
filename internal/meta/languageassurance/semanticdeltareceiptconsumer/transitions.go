package semanticdeltareceiptconsumer

import "sort"

const boundedClaimID = "gooo://semantic-delta/claim/bounded-equivalence"

func transitions(before, after projectedSource, class, reason string) []ClaimTransition {
	from, to, why := statusOpen, statusOpen, "BOUNDED_EQUIVALENCE_OPEN"
	if class == classPreserved {
		to, why = statusDischarged, "BOUNDED_EQUIVALENCE_DISCHARGED"
	}
	if class == classChanged {
		to, why = statusRefuted, "BOUNDED_EQUIVALENCE_REFUTED"
	}
	result := []ClaimTransition{{ClaimID: boundedClaimID, FromStatus: from, ToStatus: to, FromObject: before.semanticDigest, ToObject: after.semanticDigest, Stage: "adjudicate", Step: "bounded-equivalence", Reason: why}}
	if class != classChanged {
		return result
	}
	left, right := claimMap(before.claims), claimMap(after.claims)
	ids := make([]string, 0, len(left)+len(right))
	seen := make(map[string]bool, len(left)+len(right))
	for id := range left {
		ids = append(ids, id)
		seen[id] = true
	}
	for id := range right {
		if !seen[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		claim, leftOK := left[id]
		other, rightOK := right[id]
		if leftOK && rightOK && claimMeaningEqual(claim, other) {
			continue
		}
		fromObject, toObject := "", ""
		if leftOK {
			fromObject = claim.Object
		}
		if rightOK {
			toObject = other.Object
		}
		result = append(result, changedClaimTransition(id, fromObject, toObject, reason))
	}
	return result
}

func objectOf(values map[string]Claim, id string) string {
	if value, ok := values[id]; ok {
		return value.Object
	}
	return ""
}

func changedClaimTransition(id, fromObject, toObject, reason string) ClaimTransition {
	return ClaimTransition{ClaimID: id, FromStatus: statusOpen, ToStatus: statusRefuted, FromObject: fromObject, ToObject: toObject, Stage: "adjudicate", Step: "claim-transition", Reason: reason}
}
