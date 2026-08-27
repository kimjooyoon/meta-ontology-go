package semanticdeltareceiptconsumer

const boundedClaimID = "gooo://semantic-delta/claim/bounded-equivalence"

func claimLedger(before, after projectedSource, class string) ([]Claim, []ClaimTransition) {
	from, to, why := statusOpen, statusOpen, "BOUNDED_EQUIVALENCE_OPEN"
	if class == classPreserved {
		to, why = statusDischarged, "BOUNDED_EQUIVALENCE_DISCHARGED"
	}
	if class == classChanged {
		to, why = statusRefuted, "BOUNDED_EQUIVALENCE_REFUTED"
	}
	bounded := boundedClaim(after.semanticDigest, to)
	ledger := []Claim{bounded}
	result := []ClaimTransition{{ClaimID: boundedClaimID, Kind: claimKindBounded, FromStatus: from, ToStatus: to, FromObject: before.semanticDigest, ToObject: after.semanticDigest, Stage: "adjudicate", Step: "bounded-equivalence", Reason: why}}
	if class != classPreserved && class != classChanged {
		return ledger, result
	}

	left, right := claimMap(before.claims), claimMap(after.claims)
	for _, claim := range before.claims {
		ledger = append(ledger, objectObservation(claim))
	}
	for _, claim := range before.claims {
		other, found := right[claim.ID]
		status, preservationReason := statusRefuted, "BEFORE_CLAIM_NOT_PRESERVED"
		toObject := ""
		if found && claimMeaningEqual(claim, other) {
			status, preservationReason, toObject = statusDischarged, "BEFORE_CLAIM_PRESERVED", other.Object
		}
		ledger = append(ledger, preservationClaim(claim, toObject, status, preservationReason))
		result = append(result, preservationTransition(claim, toObject, status, preservationReason))
	}
	for _, claim := range after.claims {
		ledger = append(ledger, objectObservation(claim))
		if _, found := left[claim.ID]; !found {
			result = append(result, objectObservationTransition(claim))
		}
	}
	ledger = uniqueClaims(ledger)
	return ledger, result
}
