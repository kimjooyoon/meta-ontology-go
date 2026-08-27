package semanticdeltareceipt

import "sort"

func pairProposition(before, after projectedSource) string {
	return canonicalPairTargetAddress(before.semanticDigest, after.semanticDigest)
}

func unknownLedger(before, after projectedSource, stage, step, reason string) ([]Claim, []ClaimTransition) {
	bounded := boundedClaim(before, after, StatusOpen)
	bounded.Stage, bounded.Step, bounded.Reason = stage, step, reason
	ledger := []Claim{bounded}
	transitions := []ClaimTransition{{ClaimID: bounded.ID, Kind: ClaimKindBounded, FromStatus: StatusOpen, ToStatus: StatusOpen, FromObject: bounded.Object, ToObject: bounded.Object, PropositionDigest: bounded.PropositionDigest, Stage: stage, Step: step, Reason: reason}}
	for _, claim := range uniqueClaims(append(append([]Claim{}, before.claims...), after.claims...)) {
		claim.Status, claim.Stage, claim.Step, claim.Reason = StatusOpen, stage, step, reason
		ledger = append(ledger, claim)
		transitions = append(transitions, ClaimTransition{ClaimID: claim.ID, Kind: claim.Kind, FromStatus: StatusOpen, ToStatus: StatusOpen, FromObject: claim.Object, ToObject: claim.Object, PropositionDigest: claim.PropositionDigest, Stage: stage, Step: step, Reason: reason})
	}
	ledger = uniqueClaims(ledger)
	return ledger, sealTransitions(before, after, transitions)
}

func boundedClaim(before, after projectedSource, status string) Claim {
	object := pairProposition(before, after)
	normalized := normalizedProposition(ClaimKindBounded, "source-pair", "bounded-semantic-equivalence", object)
	proposition := propositionDigest(ClaimKindBounded, "source-pair", "bounded-semantic-equivalence", object)
	return Claim{ID: boundedClaimID(object, "bounded-equivalence"), ClaimTypeID: claimTypeID(ClaimKindBounded, "source-pair", "bounded-semantic-equivalence", "bounded-equivalence"), Kind: ClaimKindBounded, Subject: "source-pair", Predicate: "bounded-semantic-equivalence", Object: object, Status: status, Stage: "adjudicate", Step: "bounded-equivalence", Reason: "BOUNDED_EQUIVALENCE_LEDGER", NormalizedProposition: normalized, PropositionDigest: proposition, TargetAddress: object, TargetAddressDigest: targetAddressDigest(object), RelationRole: "bounded-equivalence", BeforeSourcePath: before.path, AfterSourcePath: after.path, BeforeSourceDigest: before.rawDigest, AfterSourceDigest: after.rawDigest, BeforeSemanticDigest: before.semanticDigest, AfterSemanticDigest: after.semanticDigest}
}

func preservationClaim(before, after Claim, afterObject, status, reason string) Claim {
	normalized := normalizedProposition(ClaimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition)
	claim := Claim{ID: preservationClaimID(before), ClaimTypeID: claimTypeID(ClaimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition), Kind: ClaimKindPreserve, Subject: before.ID, Predicate: "preserves", Object: normalized, Status: status, Stage: "adjudicate", Step: "before-claim-preservation", Reason: reason, NormalizedProposition: normalized, PropositionDigest: propositionDigest(ClaimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition), PreservationOf: before.ID, TargetAddress: before.TargetAddress, TargetAddressDigest: before.TargetAddressDigest, RelationRole: "preserves", BeforeSourcePath: before.BeforeSourcePath, AfterSourcePath: after.AfterSourcePath, BeforeSourceDigest: before.BeforeSourceDigest, AfterSourceDigest: after.AfterSourceDigest, BeforeSemanticDigest: before.BeforeSemanticDigest, AfterSemanticDigest: after.AfterSemanticDigest}
	if after.ID == "" {
		claim.AfterSourceDigest, claim.AfterSemanticDigest = "", ""
	}
	_ = afterObject
	return claim
}

func objectObservation(claim Claim) Claim {
	claim.Status, claim.Stage, claim.Step, claim.Reason = StatusDischarged, "adjudicate", "source-observation", "CANONICAL_SOURCE_OBSERVATION"
	return claim
}

func preservationTransition(before, after Claim, status, reason string) ClaimTransition {
	claimID := preservationClaimID(before)
	return ClaimTransition{ClaimID: claimID, Kind: ClaimKindPreserve, FromStatus: StatusOpen, ToStatus: status, FromObject: before.NormalizedProposition, ToObject: before.NormalizedProposition, PreservationOf: before.ID, PropositionDigest: propositionDigest(ClaimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition), Stage: "adjudicate", Step: "before-claim-preservation", Reason: reason}
}

func objectObservationTransition(claim Claim) ClaimTransition {
	return ClaimTransition{ClaimID: claim.ID, Kind: ClaimKindObject, FromStatus: StatusOpen, ToStatus: StatusDischarged, FromObject: claim.Object, ToObject: claim.Object, PropositionDigest: claim.PropositionDigest, Stage: "adjudicate", Step: "source-observation", Reason: "CANONICAL_SOURCE_OBSERVATION"}
}

func uniqueClaims(values []Claim) []Claim {
	result := append([]Claim(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
