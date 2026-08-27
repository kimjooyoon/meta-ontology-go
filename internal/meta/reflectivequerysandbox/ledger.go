package reflectivequerysandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type predicateResult struct{ To, Stage, Step, Reason, Material string }

func buildClaimTransitions(claims []claimSpec, attempts []Attempt, effects Effects, receiptMaterial string) []ClaimTransition {
	transitions := make([]ClaimTransition, 0, len(claims)*2)
	previous := ""
	sequence := 1
	for _, claim := range claims {
		registration := ClaimTransition{
			Sequence: sequence, ClaimID: claim.ID, Class: claim.Class, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation,
			PriorState: claim.PriorState, EvidenceAttempt: claim.EvidenceAttempt, PredicateID: claim.PredicateID, Producer: ProducerName, Consumer: ConsumerName,
			Stage: "DECLARE", Step: "register-source-claim", Reason: "CLAIM_PRIOR_STATE_OBSERVED", From: "UNRECORDED", To: claim.PriorState, PreviousDigest: previous,
			ObservedMaterialDigest: semantic.StableHashString(claim.ID + "|" + claim.PredicateID + "|" + claim.PriorState),
		}
		registration.Digest = digestTransition(registration)
		transitions = append(transitions, registration)
		previous = registration.Digest

		result := evaluateClaim(claim, attempts, effects, receiptMaterial)
		if claim.PredicateID == "claim-ledger-chained" {
			// This predicate is discharged only after the complete persistent chain is
			// built and independently validated, never by a normal query attempt.
			result = predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "verify-complete-transition-chain", Reason: "CHAIN_PENDING"}
		}
		observation := registration
		observation.Sequence = sequence + 1
		observation.Stage, observation.Step, observation.Reason = result.Stage, result.Step, result.Reason
		observation.From, observation.To, observation.PreviousDigest = claim.PriorState, result.To, previous
		observation.ObservedMaterialDigest = result.Material
		observation.Digest = digestTransition(observation)
		transitions = append(transitions, observation)
		previous = observation.Digest
		sequence += 2
	}

	chainDigest := transitionChainDigest(transitions)
	for index := range transitions {
		if transitions[index].PredicateID == "claim-ledger-chained" && index%2 == 1 {
			transitions[index].To = "DISCHARGED"
			transitions[index].Stage = "OBSERVE"
			transitions[index].Step = "verify-complete-transition-chain"
			transitions[index].Reason = "COMPLETE_TRANSITION_CHAIN_VERIFIED"
			transitions[index].ObservedMaterialDigest = chainDigest
		}
	}
	resignTransitions(transitions)
	return transitions
}

func evaluateClaim(claim claimSpec, attempts []Attempt, effects Effects, receiptMaterial string) predicateResult {
	if claim.PredicateID == "claim-ledger-chained" {
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "verify-complete-transition-chain", Reason: "CHAIN_PENDING"}
	}
	attempt, ok := findAttempt(attempts, claim.EvidenceAttempt)
	if !ok {
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "resolve-missing-evidence", Reason: "EVIDENCE_MISSING"}
	}
	contradiction := func(reason string) predicateResult {
		return predicateResult{To: "REFUTED", Stage: "REFUTE", Step: "predicate-contradiction", Reason: reason, Material: attempt.ObservedMaterialDigest}
	}
	observationError := func() predicateResult {
		return predicateResult{To: claim.PriorState, Stage: attempt.Stage, Step: attempt.Step, Reason: attempt.Reason, Material: attempt.ObservedMaterialDigest}
	}
	exact := attempt.Decision == "PASS" && attempt.Resolution == "EXACT" && attempt.MatchedFacts == 1
	sameSemantic := attempt.SemanticDigestBefore != "" && attempt.SemanticDigestBefore == attempt.SemanticDigestAfter
	sameGraph := attempt.GraphDigestBefore != "" && attempt.GraphDigestBefore == attempt.GraphDigestAfter
	switch claim.PredicateID {
	case "query-relation-exact":
		if exact {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "evaluate-query-relation", Reason: "EXACT_RELATION_MATCH", Material: attempt.GraphDigestAfter}
		}
		if attempt.Decision == "REFUTED" {
			return contradiction("QUERY_RELATION_CONTRADICTION")
		}
		return observationError()
	case "semantic-digest-equal":
		if sameSemantic {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "compare-semantic-digest", Reason: "SEMANTIC_DIGEST_EQUAL", Material: semantic.StableHashString(attempt.SemanticDigestBefore + "|" + attempt.SemanticDigestAfter)}
		}
		if attempt.SemanticDigestBefore != "" && attempt.SemanticDigestAfter != "" {
			return contradiction("SEMANTIC_DIGEST_CHANGED")
		}
		return observationError()
	case "graph-digest-equal":
		if sameGraph {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "compare-graph-digest", Reason: "GRAPH_DIGEST_EQUAL", Material: attempt.GraphDigestAfter}
		}
		if attempt.GraphDigestBefore != "" && attempt.GraphDigestAfter != "" {
			return contradiction("GRAPH_DIGEST_CHANGED")
		}
		return observationError()
	case "query-projection-stable":
		if exact && sameGraph {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-query-projection", Reason: "QUERY_MATCH_AND_GRAPH_STABLE", Material: attempt.GraphDigestAfter}
		}
		if attempt.Decision == "REFUTED" {
			return contradiction("QUERY_PROJECTION_CONTRADICTION")
		}
		return observationError()
	case "receipt-observation-digest-verified":
		if attempt.ObservedMaterialDigest != "" && receiptMaterial != "" && attempt.ObservedMaterialDigest == receiptMaterial {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-receipt-material-digest", Reason: "RECEIPT_MATERIAL_DIGEST_VERIFIED", Material: receiptMaterial}
		}
		if attempt.ObservedMaterialDigest != "" && receiptMaterial != "" {
			return contradiction("RECEIPT_MATERIAL_DIGEST_MISMATCH")
		}
		return observationError()
	case "unknown-subject-preserved":
		if attempt.Decision == "UNKNOWN" && attempt.Resolution == "LOWER_RESOLUTION" && attempt.Stage == "UNKNOWN" && attempt.Step == "resolve-unknown-subject" && attempt.Reason == "UNKNOWN_TARGET" {
			return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "retain-open-on-unknown", Reason: "UNKNOWN_PRESERVED", Material: attempt.GraphDigestAfter}
		}
		if attempt.Decision == "REFUTED" {
			return contradiction("UNKNOWN_SUBJECT_BOUNDARY_CONTRADICTION")
		}
		return observationError()
	case "immutable-id-patch-rejected":
		if immutableRejection(attempt) {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-immutable-id-patch-rejection", Reason: "IMMUTABLE_ID_PATCH_REJECTED", Material: attempt.OriginalGraphDigestAfter}
		}
		if attempt.Decision == "REFUTED" || attempt.APIOutcome == "ACCEPTED" {
			return contradiction("IMMUTABLE_ID_PATCH_ACCEPTED")
		}
		return observationError()
	case "immutable-id-patch-accepted-false":
		if immutableRejection(attempt) && !effects.ImmutableIDPatchAccepted && effects.MutationOutcome == "REJECTED" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-scoped-immutable-id-fact", Reason: "IMMUTABLE_ID_PATCH_ACCEPTED_FALSE", Material: semantic.StableHashString(attempt.OriginalGraphDigestAfter + "|immutable_id_patch_accepted=false")}
		}
		if effects.ImmutableIDPatchAccepted || attempt.APIOutcome == "ACCEPTED" || attempt.Decision == "REFUTED" {
			return contradiction("IMMUTABLE_ID_PATCH_ACCEPTED")
		}
		return observationError()
	case "net-repository-status-unchanged":
		if reflectRepositoryNet(effects) && attempt.Decision == "PASS" && attempt.Resolution == "EXACT" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-net-repository-status", Reason: "NET_REPOSITORY_STATUS_UNCHANGED", Material: attempt.ObservedMaterialDigest}
		}
		if effects.RepositoryEvidenceAvailable && len(effects.NetRepositoryChanges) > 0 {
			return contradiction("NET_REPOSITORY_STATUS_CHANGED")
		}
		return observationError()
	default:
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "reject-unknown-predicate", Reason: "PREDICATE_NOT_ALLOWED"}
	}
}

func immutableRejection(attempt Attempt) bool {
	return attempt.Decision == "DENIED" && attempt.Resolution == "EXACT_REJECTION" && attempt.APIOutcome == "REJECTED" && attempt.APIErrorCode == string(semantic.PatchImmutableField) && attempt.MutationField == "id" && attempt.GraphDigestBefore != "" && attempt.GraphDigestBefore == attempt.OriginalGraphDigestAfter && attempt.SemanticDigestBefore == attempt.OriginalSemanticDigestAfter && attempt.ReturnedGraphDigest == ""
}

func reflectRepositoryNet(effects Effects) bool {
	return effects.RepositoryEvidenceAvailable && effects.RepositoryObservation == "net_repository_status_unchanged" && effects.RepositoryStatusBefore != nil && effects.RepositoryStatusAfter != nil && len(effects.NetRepositoryChanges) == 0 && stringSliceEqual(effects.RepositoryStatusBefore, effects.RepositoryStatusAfter)
}

func stringSliceEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func findAttempt(attempts []Attempt, id string) (Attempt, bool) {
	for _, attempt := range attempts {
		if attempt.ID == id {
			return attempt, true
		}
	}
	return Attempt{}, false
}

type receiptMaterial struct {
	Source   Snapshot  `json:"source"`
	Attempts []Attempt `json:"attempts"`
	Effects  Effects   `json:"effects"`
}

func receiptMaterialDigest(source Snapshot, attempts []Attempt, effects Effects) string {
	copyAttempts := append([]Attempt(nil), attempts...)
	for index := range copyAttempts {
		copyAttempts[index].ObservedMaterialDigest = ""
	}
	return hashJSON(receiptMaterial{Source: source, Attempts: copyAttempts, Effects: effects})
}

func validateTransitionChain(transitions []ClaimTransition) (string, error) {
	if len(transitions) == 0 || len(transitions)%2 != 0 {
		return "", fmt.Errorf("transition chain has invalid length %d", len(transitions))
	}
	previous := ""
	for index, transition := range transitions {
		if transition.Sequence != index+1 || transition.PreviousDigest != previous || transition.Digest != digestTransition(transition) {
			return "", fmt.Errorf("transition chain link %d is invalid", index+1)
		}
		if transition.ClaimID == "" || transition.Class == "" || transition.ProofChoice == "" || transition.PredicateID == "" || transition.PriorState != "OPEN" || transition.Producer != ProducerName || transition.Consumer != ConsumerName {
			return "", fmt.Errorf("transition %d lacks complete claim identity", index+1)
		}
		if index%2 == 0 {
			if transition.From != "UNRECORDED" || transition.To != transition.PriorState || transition.Stage != "DECLARE" {
				return "", fmt.Errorf("claim registration %d is invalid", index+1)
			}
		} else {
			registration := transitions[index-1]
			if transition.ClaimID != registration.ClaimID || transition.Class != registration.Class || transition.ProofChoice != registration.ProofChoice || transition.MetaOperation != registration.MetaOperation || transition.PriorState != registration.PriorState || transition.EvidenceAttempt != registration.EvidenceAttempt || transition.PredicateID != registration.PredicateID || transition.From != transition.PriorState {
				return "", fmt.Errorf("claim transition %d is not bound to its registration", index+1)
			}
		}
		previous = transition.Digest
	}
	return transitionChainDigest(transitions), nil
}

func transitionChainDigest(transitions []ClaimTransition) string {
	canonical := append([]ClaimTransition(nil), transitions...)
	for index := range canonical {
		canonical[index].Digest = ""
		canonical[index].PreviousDigest = ""
		canonical[index].ObservedMaterialDigest = ""
		if index%2 == 1 && canonical[index].PredicateID == "claim-ledger-chained" {
			canonical[index].To = "DISCHARGED"
			canonical[index].Stage = "OBSERVE"
			canonical[index].Step = "verify-complete-transition-chain"
			canonical[index].Reason = "COMPLETE_TRANSITION_CHAIN_VERIFIED"
		}
	}
	return hashJSON(canonical)
}

func resignTransitions(transitions []ClaimTransition) {
	previous := ""
	for index := range transitions {
		transitions[index].Sequence = index + 1
		transitions[index].PreviousDigest = previous
		transitions[index].Digest = digestTransition(transitions[index])
		previous = transitions[index].Digest
	}
}

func observationDigest(value Observation) string {
	value.Digest = ""
	value.ProvisionalDigest = ""
	return hashJSON(value)
}

func digestTransition(value ClaimTransition) string {
	value.Digest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func hashJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
