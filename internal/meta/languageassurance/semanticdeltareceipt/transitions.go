package semanticdeltareceipt

const boundedEquivalenceClaim = "gooo://semantic-delta/claim/bounded-equivalence"

func claimLedger(before, after projectedSource, class, reason string) ([]Claim, []ClaimTransition) {
	from, to, why := StatusOpen, StatusOpen, "BOUNDED_EQUIVALENCE_OPEN"
	if class == ClassPreserved {
		to, why = StatusDischarged, "BOUNDED_EQUIVALENCE_DISCHARGED"
	}
	if class == ClassChanged {
		to, why = StatusRefuted, "BOUNDED_EQUIVALENCE_REFUTED"
	}
	bounded := boundedClaim(after.semanticDigest, to)
	ledger := []Claim{bounded}
	result := []ClaimTransition{{ClaimID: boundedEquivalenceClaim, Kind: ClaimKindBounded, FromStatus: from, ToStatus: to, FromObject: before.semanticDigest, ToObject: after.semanticDigest, Stage: "adjudicate", Step: "bounded-equivalence", Reason: why}}
	if class != ClassPreserved && class != ClassChanged {
		return ledger, result
	}

	left, right := claimMap(before.claims), claimMap(after.claims)
	for _, claim := range before.claims {
		ledger = append(ledger, objectObservation(claim))
	}
	for _, claim := range before.claims {
		other, found := right[claim.ID]
		status, preservationReason := StatusRefuted, "BEFORE_CLAIM_NOT_PRESERVED"
		toObject := ""
		if found && claimMeaningEqual(claim, other) {
			status, preservationReason, toObject = StatusDischarged, "BEFORE_CLAIM_PRESERVED", other.Object
		}
		ledger = append(ledger, preservationClaim(claim, toObject, status, preservationReason))
		result = append(result, preservationTransition(claim, toObject, status, preservationReason))
	}
	for _, claim := range after.claims {
		observed := objectObservation(claim)
		ledger = append(ledger, observed)
		if _, found := left[claim.ID]; !found {
			result = append(result, objectObservationTransition(claim))
		}
	}
	ledger = uniqueClaims(ledger)
	return ledger, result
}
