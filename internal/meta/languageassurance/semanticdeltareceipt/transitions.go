package semanticdeltareceipt

import "strings"

func claimLedger(before, after projectedSource, class string, reasons ...string) ([]Claim, []ClaimTransition) {
	from, to, why := StatusOpen, StatusOpen, "BOUNDED_EQUIVALENCE_OPEN"
	if class == ClassPreserved {
		to, why = StatusDischarged, "BOUNDED_EQUIVALENCE_DISCHARGED"
	}
	if class == ClassChanged {
		to, why = StatusRefuted, "BOUNDED_EQUIVALENCE_REFUTED"
	}
	bounded := boundedClaim(before, after, to)
	ledger := []Claim{bounded}
	transitions := []ClaimTransition{{ClaimID: bounded.ID, Kind: ClaimKindBounded, FromStatus: from, ToStatus: to, FromObject: bounded.Object, ToObject: bounded.Object, PropositionDigest: bounded.PropositionDigest, Stage: "adjudicate", Step: "bounded-equivalence", Reason: why}}
	if class != ClassPreserved && class != ClassChanged {
		return sealLedger(before, after, ledger, transitions)
	}

	for _, claim := range before.claims {
		observed := objectObservation(claim)
		ledger = append(ledger, observed)
		transitions = append(transitions, objectObservationTransition(observed))
	}
	right := claimsBySlot(after.claims)
	for _, claim := range before.claims {
		matches := right[claim.Subject+"\x00"+claim.Predicate]
		status, preservationReason := StatusRefuted, "BEFORE_CLAIM_NOT_PRESERVED"
		var other Claim
		if len(matches) == 1 && claimMeaningEqual(claim, matches[0]) {
			other, status, preservationReason = matches[0], StatusDischarged, "BEFORE_CLAIM_PRESERVED"
		}
		ledger = append(ledger, preservationClaim(claim, other, other.Object, status, preservationReason))
		transitions = append(transitions, preservationTransition(claim, other, status, preservationReason))
	}
	for _, claim := range after.claims {
		observed := objectObservation(claim)
		ledger = append(ledger, observed)
		transitions = append(transitions, objectObservationTransition(observed))
	}
	return sealLedger(before, after, uniqueClaims(ledger), transitions)
}

func sealLedger(before, after projectedSource, ledger []Claim, transitions []ClaimTransition) ([]Claim, []ClaimTransition) {
	return ledger, sealTransitions(before, after, transitions)
}

func sealTransitions(before, after projectedSource, transitions []ClaimTransition) []ClaimTransition {
	previous := ""
	evidence := digestValue(strings.Join([]string{before.path, before.rawDigest, before.semanticDigest, after.path, after.rawDigest, after.semanticDigest}, "\x00"))
	for index := range transitions {
		if transitions[index].PropositionDigest == "" {
			transitions[index].PropositionDigest = digestValue(strings.Join([]string{transitions[index].Kind, transitions[index].FromObject, transitions[index].ToObject}, "\x00"))
		}
		transitions[index].EvidenceDigest = digestValue(strings.Join([]string{evidence, transitions[index].PropositionDigest}, "\x00"))
		transitions[index].PreviousEventDigest = previous
		transitions[index].EventID = "gooo://semantic-delta/event/" + digestValue(strings.Join([]string{transitions[index].ClaimID, transitions[index].EvidenceDigest, previous}, "\x00"))[len("sha256:"):]
		copy := transitions[index]
		copy.TransitionDigest = ""
		transitions[index].TransitionDigest = digestValue(copy)
		previous = transitions[index].TransitionDigest
	}
	return transitions
}
