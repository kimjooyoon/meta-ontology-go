package reflectivequerysandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type predicateResult struct{ To, Stage, Step, Reason, Material string }

func buildClaimTransitions(claims []claimSpec, attempts []Attempt, effects Effects, receiptMaterial string) []ClaimTransition {
	transitions := make([]ClaimTransition, 0, len(claims)*2)
	previous := ""
	sequence := 1
	for _, claim := range claims {
		registration := ClaimTransition{Sequence: sequence, ClaimID: claim.ID, Class: claim.Class, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation, PriorState: claim.PriorState, EvidenceAttempt: claim.EvidenceAttempt, PredicateID: claim.PredicateID, Producer: ProducerName, Consumer: ConsumerName, Stage: "DECLARE", Step: "register-source-claim", Reason: "CLAIM_PRIOR_STATE_OBSERVED", From: "UNRECORDED", To: claim.PriorState, PreviousDigest: previous}
		registration.ObservedMaterialDigest = semantic.StableHashString(claim.ID + "|" + claim.PredicateID + "|" + claim.PriorState)
		registration.Digest = digestTransition(registration)
		transitions = append(transitions, registration)
		previous = registration.Digest
		result := evaluateClaim(claim, attempts, effects, receiptMaterial)
		observation := registration
		observation.Sequence = sequence + 1
		observation.Stage, observation.Step, observation.Reason = result.Stage, result.Step, result.Reason
		observation.From, observation.To, observation.PreviousDigest = claim.PriorState, result.To, previous
		if claim.PredicateID == "claim-ledger-chained" && observation.PreviousDigest != registration.Digest {
			result = predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "verify-claim-ledger-chain", Reason: "CLAIM_LEDGER_CHAIN_BROKEN", Material: registration.Digest}
			observation.Stage, observation.Step, observation.Reason = result.Stage, result.Step, result.Reason
			observation.To, observation.ObservedMaterialDigest = result.To, result.Material
		}
		observation.ObservedMaterialDigest = result.Material
		observation.Digest = digestTransition(observation)
		transitions = append(transitions, observation)
		previous = observation.Digest
		sequence += 2
	}
	return transitions
}

func evaluateClaim(claim claimSpec, attempts []Attempt, effects Effects, receiptMaterial string) predicateResult {
	attempt, ok := findAttempt(attempts, claim.EvidenceAttempt)
	if !ok {
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "resolve-missing-evidence", Reason: "EVIDENCE_MISSING"}
	}
	contradiction := func(reason string) predicateResult {
		return predicateResult{To: "REFUTED", Stage: "REFUTE", Step: "predicate-contradiction", Reason: reason, Material: attempt.ObservedMaterialDigest}
	}
	observationError := func() predicateResult {
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "predicate-observation-error", Reason: attempt.Reason, Material: attempt.ObservedMaterialDigest}
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
	case "claim-ledger-chained":
		if attempt.ObservedMaterialDigest != "" && attempt.Decision == "PASS" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-claim-ledger-evidence", Reason: "CLAIM_LEDGER_EVIDENCE_PRESENT", Material: attempt.ObservedMaterialDigest}
		}
		if attempt.Decision == "REFUTED" {
			return contradiction("CLAIM_LEDGER_EVIDENCE_CONTRADICTION")
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
	case "immutable-mutation-rejected":
		if immutableRejection(attempt) {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-immutable-rejection", Reason: "IMMUTABLE_FIELD_REJECTED", Material: attempt.OriginalGraphDigestAfter}
		}
		if attempt.Decision == "REFUTED" || attempt.APIOutcome == "ACCEPTED" {
			return contradiction("IMMUTABLE_MUTATION_BOUNDARY_VIOLATED")
		}
		return observationError()
	case "mutation-boundary-rejected":
		if immutableRejection(attempt) && !effects.MutationAuthority && effects.MutationOutcome == "REJECTED" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-mutation-boundary", Reason: "TYPED_REJECTION_AND_GRAPH_UNCHANGED", Material: attempt.OriginalGraphDigestAfter}
		}
		if effects.MutationAuthority || attempt.APIOutcome == "ACCEPTED" || attempt.Decision == "REFUTED" {
			return contradiction("MUTATION_AUTHORITY_OBSERVED")
		}
		return observationError()
	case "net-repository-changes-empty":
		if reflectRepositoryNet(effects) && attempt.Decision == "PASS" && attempt.Resolution == "EXACT" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-net-repository-status", Reason: "NET_REPOSITORY_CHANGES_EMPTY", Material: attempt.ObservedMaterialDigest}
		}
		if len(effects.NetRepositoryChanges) > 0 {
			return contradiction("NET_REPOSITORY_CHANGES_OBSERVED")
		}
		return observationError()
	default:
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "reject-unknown-predicate", Reason: "PREDICATE_NOT_ALLOWED"}
	}
}

func immutableRejection(attempt Attempt) bool {
	return attempt.Decision == "DENIED" && attempt.Resolution == "EXACT_REJECTION" && attempt.APIOutcome == "REJECTED" && attempt.APIErrorCode == string(semantic.PatchImmutableField) && attempt.GraphDigestBefore != "" && attempt.GraphDigestBefore == attempt.OriginalGraphDigestAfter && attempt.SemanticDigestBefore == attempt.OriginalSemanticDigestAfter && attempt.ReturnedGraphDigest == ""
}

func reflectRepositoryNet(effects Effects) bool {
	return effects.RepositoryStatusBefore != nil && effects.RepositoryStatusAfter != nil && len(effects.NetRepositoryChanges) == 0 && stringSliceEqual(effects.RepositoryStatusBefore, effects.RepositoryStatusAfter)
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

func digestTransition(value ClaimTransition) string {
	value.Digest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func SealObservation(value Observation) Observation {
	value.Digest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	value.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return value
}

func hashJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
