package semanticdeltareceiptconsumer

import (
	"sort"
	"strings"
)

func pairProposition(before, after projectedSource) string {
	return strings.Join([]string{before.path, after.path, before.rawDigest, after.rawDigest, before.semanticDigest, after.semanticDigest}, "\x00")
}

func unknownLedger(before, after projectedSource, stage, step, reason string) ([]Claim, []ClaimTransition) {
	bounded := boundedClaim(before, after, statusOpen)
	bounded.Stage, bounded.Step, bounded.Reason = stage, step, reason
	ledger := []Claim{bounded}
	transitions := []ClaimTransition{{ClaimID: bounded.ID, Kind: claimKindBounded, FromStatus: statusOpen, ToStatus: statusOpen, FromObject: bounded.Object, ToObject: bounded.Object, PropositionDigest: bounded.PropositionDigest, Stage: stage, Step: step, Reason: reason}}
	for _, claim := range uniqueClaims(append(append([]Claim{}, before.claims...), after.claims...)) {
		claim.Status, claim.Stage, claim.Step, claim.Reason = statusOpen, stage, step, reason
		ledger = append(ledger, claim)
		transitions = append(transitions, ClaimTransition{ClaimID: claim.ID, Kind: claim.Kind, FromStatus: statusOpen, ToStatus: statusOpen, FromObject: claim.Object, ToObject: claim.Object, PropositionDigest: claim.PropositionDigest, Stage: stage, Step: step, Reason: reason})
	}
	ledger = uniqueClaims(ledger)
	return ledger, sealTransitions(before, after, transitions)
}

func boundedClaim(before, after projectedSource, status string) Claim {
	object := pairProposition(before, after)
	normalized := normalizedProposition(claimKindBounded, "source-pair", "bounded-semantic-equivalence", object)
	proposition := propositionDigest(claimKindBounded, "source-pair", "bounded-semantic-equivalence", object)
	return Claim{ID: "gooo://semantic-delta/claim/bounded-equivalence/" + proposition[len("sha256:"):], ClaimTypeID: claimTypeID(claimKindBounded, "source-pair", "bounded-semantic-equivalence", "bounded-equivalence"), Kind: claimKindBounded, Subject: "source-pair", Predicate: "bounded-semantic-equivalence", Object: object, Status: status, Stage: "adjudicate", Step: "bounded-equivalence", Reason: "BOUNDED_EQUIVALENCE_LEDGER", NormalizedProposition: normalized, PropositionDigest: proposition, TargetAddress: before.path + "->" + after.path, BeforeSourceDigest: before.rawDigest, AfterSourceDigest: after.rawDigest, BeforeSemanticDigest: before.semanticDigest, AfterSemanticDigest: after.semanticDigest}
}

func preservationClaim(before, after Claim, afterObject, status, reason string) Claim {
	normalized := normalizedProposition(claimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition)
	claim := Claim{ID: preservationClaimID(before, after, afterObject), ClaimTypeID: claimTypeID(claimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition), Kind: claimKindPreserve, Subject: before.ID, Predicate: "preserves", Object: normalized, Status: status, Stage: "adjudicate", Step: "before-claim-preservation", Reason: reason, NormalizedProposition: normalized, PropositionDigest: propositionDigest(claimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition), PreservationOf: before.ID, TargetAddress: before.TargetAddress, BeforeSourceDigest: before.BeforeSourceDigest, AfterSourceDigest: after.AfterSourceDigest, BeforeSemanticDigest: before.BeforeSemanticDigest, AfterSemanticDigest: after.AfterSemanticDigest}
	if after.ID == "" {
		claim.AfterSourceDigest, claim.AfterSemanticDigest = "", ""
	}
	_ = afterObject
	return claim
}

func objectObservation(claim Claim) Claim {
	claim.Status, claim.Stage, claim.Step, claim.Reason = statusDischarged, "adjudicate", "source-observation", "CANONICAL_SOURCE_OBSERVATION"
	return claim
}

func preservationTransition(before, after Claim, status, reason string) ClaimTransition {
	claimID := preservationClaimID(before, after, after.Object)
	return ClaimTransition{ClaimID: claimID, Kind: claimKindPreserve, FromStatus: statusOpen, ToStatus: status, FromObject: before.NormalizedProposition, ToObject: before.NormalizedProposition, PreservationOf: before.ID, PropositionDigest: propositionDigest(claimKindPreserve, before.ClaimTypeID, "preserves", before.NormalizedProposition), Stage: "adjudicate", Step: "before-claim-preservation", Reason: reason}
}

func objectObservationTransition(claim Claim) ClaimTransition {
	return ClaimTransition{ClaimID: claim.ID, Kind: claimKindObject, FromStatus: statusOpen, ToStatus: statusDischarged, FromObject: claim.Object, ToObject: claim.Object, PropositionDigest: claim.PropositionDigest, Stage: "adjudicate", Step: "source-observation", Reason: "CANONICAL_SOURCE_OBSERVATION"}
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
