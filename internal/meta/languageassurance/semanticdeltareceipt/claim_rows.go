package semanticdeltareceipt

import "sort"

func unknownLedger(stage, step, reason string) ([]Claim, []ClaimTransition) {
	claim := boundedClaim("", StatusOpen)
	claim.Stage, claim.Step, claim.Reason = stage, step, reason
	return []Claim{claim}, []ClaimTransition{{ClaimID: boundedEquivalenceClaim, Kind: ClaimKindBounded, FromStatus: StatusOpen, ToStatus: StatusOpen, Stage: stage, Step: step, Reason: reason}}
}

func boundedClaim(afterDigest, status string) Claim {
	return Claim{ID: boundedEquivalenceClaim, Kind: ClaimKindBounded, Subject: "source-pair", Predicate: "bounded-semantic-equivalence", Object: afterDigest, Status: status, Stage: "adjudicate", Step: "bounded-equivalence", Reason: "BOUNDED_EQUIVALENCE_LEDGER", PropositionDigest: propositionDigest(ClaimKindBounded, "source-pair", "bounded-semantic-equivalence", "after")}
}

func preservationClaim(before Claim, afterObject, status, reason string) Claim {
	digest := propositionDigest(ClaimKindPreserve, before.ID, "preserves", before.PropositionDigest)
	return Claim{ID: preservationClaimID(before), Kind: ClaimKindPreserve, Subject: before.ID, Predicate: "preserves", Object: afterObject, Status: status, Stage: "adjudicate", Step: "before-claim-preservation", Reason: reason, PropositionDigest: digest, PreservationOf: before.ID}
}

func objectObservation(claim Claim) Claim {
	claim.Status, claim.Stage, claim.Step, claim.Reason = StatusDischarged, "adjudicate", "source-observation", "CANONICAL_SOURCE_OBSERVATION"
	return claim
}

func preservationTransition(before Claim, afterObject, status, reason string) ClaimTransition {
	return ClaimTransition{ClaimID: preservationClaimID(before), Kind: ClaimKindPreserve, FromStatus: StatusOpen, ToStatus: status, FromObject: before.Object, ToObject: afterObject, PreservationOf: before.ID, Stage: "adjudicate", Step: "before-claim-preservation", Reason: reason}
}

func objectObservationTransition(claim Claim) ClaimTransition {
	return ClaimTransition{ClaimID: claim.ID, Kind: ClaimKindObject, FromStatus: StatusOpen, ToStatus: StatusDischarged, ToObject: claim.Object, Stage: "adjudicate", Step: "source-observation", Reason: "CANONICAL_SOURCE_OBSERVATION"}
}

func uniqueClaims(values []Claim) []Claim {
	byID := make(map[string]Claim, len(values))
	for _, value := range values {
		byID[value.ID] = value
	}
	result := make([]Claim, 0, len(byID))
	for _, value := range byID {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
