package reflectivequerysandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func buildClaimTransitions() []ClaimTransition {
	specs := claimSpecs()
	transitions := make([]ClaimTransition, 0, len(specs)*2)
	previous := ""
	sequence := 1
	for _, spec := range specs {
		registration := ClaimTransition{Sequence: sequence, ClaimID: spec.ID, Class: spec.Class, ProofChoice: spec.ProofChoice, MetaOperation: spec.MetaOperation, Producer: spec.Producer, Consumer: spec.Consumer, Stage: "DECLARE", Step: "register-denominator-claim", Reason: "CLAIM_REGISTERED", From: "UNRECORDED", To: "OPEN", PreviousDigest: previous}
		registration.Digest = digestTransition(registration)
		transitions = append(transitions, registration)
		previous = registration.Digest
		sequence++
		observation := registration
		observation.Sequence = sequence
		observation.Stage, observation.Step = "OBSERVE", "evaluate-read-only-boundary"
		observation.Reason, observation.From, observation.To = "OBSERVATION_DISCHARGED", "OPEN", "DISCHARGED"
		observation.PreviousDigest = previous
		observation.Digest = digestTransition(observation)
		transitions = append(transitions, observation)
		previous = observation.Digest
		sequence++
	}
	return transitions
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
