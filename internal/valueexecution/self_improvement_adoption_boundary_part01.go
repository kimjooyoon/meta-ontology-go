package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementAdoptionBoundaryObservationSchema = "gooo.self-improvement.adoption-boundary.v1"

const (
	SelfImprovementAdoptionBoundaryAdoptable = "ADOPTABLE"
	SelfImprovementAdoptionBoundaryRejected  = "REJECTED"
	SelfImprovementAdoptionBoundaryUnknown   = "UNKNOWN"
)

type SelfImprovementAdoptionBoundaryInput struct {
	CandidateDigest string
	Contract        SelfImprovementContractPreservationObservation
	Reverse         SelfImprovementReverseClosureObservation
}

type SelfImprovementAdoptionBoundaryObservation struct {
	Schema             string `json:"schema"`
	CandidateDigest    string `json:"candidateDigest"`
	ContractStatus     string `json:"contractStatus"`
	ReverseStatus      string `json:"reverseStatus"`
	Status             string `json:"status"`
	Reason             string `json:"reason"`
	ContractPreserved  bool   `json:"contractPreserved"`
	ReverseObserved    bool   `json:"reverseObserved"`
	AdoptionAuthorized bool   `json:"adoptionAuthorized"`
	NonAuthorizing     bool   `json:"nonAuthorizing"`
	Digest             string `json:"digest"`
}

func ObserveSelfImprovementAdoptionBoundary(
	input SelfImprovementAdoptionBoundaryInput,
) SelfImprovementAdoptionBoundaryObservation {
	observation := SelfImprovementAdoptionBoundaryObservation{
		Schema:             SelfImprovementAdoptionBoundaryObservationSchema,
		CandidateDigest:    input.CandidateDigest,
		ContractStatus:     input.Contract.Status,
		ReverseStatus:      input.Reverse.Status,
		ContractPreserved:  input.Contract.ContractPreserved,
		ReverseObserved:   input.Reverse.ReverseObservation,
		AdoptionAuthorized: false,
		NonAuthorizing:     true,
	}

	switch {
	case input.CandidateDigest == "":
		observation.Status = SelfImprovementAdoptionBoundaryUnknown
		observation.Reason = "CANDIDATE_DIGEST_MISSING"
	case input.Contract.Status == SelfImprovementContractPreservationRejected:
		observation.Status = SelfImprovementAdoptionBoundaryRejected
		observation.Reason = "CONTRACT_PRESERVATION_REJECTED"
	case input.Reverse.Status == SelfImprovementReverseClosureNotApplied:
		observation.Status = SelfImprovementAdoptionBoundaryRejected
		observation.Reason = "REVERSE_CLOSURE_NOT_APPLIED"
	case input.Contract.Status != SelfImprovementContractPreservationAccepted ||
		input.Reverse.Status == SelfImprovementReverseClosureUnknown:
		observation.Status = SelfImprovementAdoptionBoundaryUnknown
		observation.Reason = "ADOPTION_EVIDENCE_INCOMPLETE"
	case input.Reverse.Status != SelfImprovementReverseClosureCompleted:
		observation.Status = SelfImprovementAdoptionBoundaryUnknown
		observation.Reason = "REVERSE_CLOSURE_INCOMPLETE"
	default:
		observation.Status = SelfImprovementAdoptionBoundaryAdoptable
		observation.Reason = "EVIDENCE_COMPLETE_ADOPTION_PENDING"
	}

	observation.Digest = digestSelfImprovementAdoptionBoundary(observation)
	return observation
}

func digestSelfImprovementAdoptionBoundary(
	observation SelfImprovementAdoptionBoundaryObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}