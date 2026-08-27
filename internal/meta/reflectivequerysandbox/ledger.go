package reflectivequerysandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func buildClaimTransitions(claims []claimSpec, attempts []Attempt) []ClaimTransition {
	transitions := make([]ClaimTransition, 0, len(claims)*2)
	previous := ""
	sequence := 1
	for _, claim := range claims {
		registration := ClaimTransition{
			Sequence: sequence, ClaimID: claim.ID, Class: claim.Class, ProofChoice: claim.ProofChoice,
			MetaOperation: claim.MetaOperation, PriorState: claim.PriorState, EvidenceAttempt: claim.EvidenceAttempt,
			Producer: ProducerName, Consumer: ConsumerName, Stage: "DECLARE", Step: "register-source-claim",
			Reason: "CLAIM_PRIOR_STATE_OBSERVED", From: "UNRECORDED", To: claim.PriorState, PreviousDigest: previous,
		}
		registration.Digest = digestTransition(registration)
		transitions = append(transitions, registration)
		previous = registration.Digest
		sequence++

		to, stage, step, reason := claim.PriorState, "OBSERVE", "retain-prior-state", "NO_ATTEMPT_EVIDENCE"
		if attempt, ok := findAttempt(attempts, claim.EvidenceAttempt); ok {
			switch attempt.Decision {
			case "PASS", "DENIED":
				to, stage, step, reason = "DISCHARGED", "OBSERVE", "discharge-from-attempt-evidence", "ATTEMPT_EVIDENCE_MATCH"
			case "UNKNOWN":
				to, stage, step, reason = claim.PriorState, "RESOLVE", "retain-open-on-unknown", "UNKNOWN_PRESERVED"
			case "REFUTED":
				to, stage, step, reason = "REFUTED", "REFUTE", "mark-boundary-violation", "BOUNDARY_VIOLATION"
			}
		}
		observation := registration
		observation.Sequence = sequence
		observation.Stage, observation.Step, observation.Reason = stage, step, reason
		observation.From, observation.To, observation.PreviousDigest = claim.PriorState, to, previous
		observation.Digest = digestTransition(observation)
		transitions = append(transitions, observation)
		previous = observation.Digest
		sequence++
	}
	return transitions
}

func findAttempt(attempts []Attempt, id string) (Attempt, bool) {
	for _, attempt := range attempts {
		if attempt.ID == id {
			return attempt, true
		}
	}
	return Attempt{}, false
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
