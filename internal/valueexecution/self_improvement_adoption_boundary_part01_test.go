package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementAdoptionBoundaryKeepsAdoptionPending(t *testing.T) {
	observation := ObserveSelfImprovementAdoptionBoundary(
		SelfImprovementAdoptionBoundaryInput{
			CandidateDigest: "candidate-digest",
			Contract: SelfImprovementContractPreservationObservation{
				Status:            SelfImprovementContractPreservationAccepted,
				ContractPreserved: true,
			},
			Reverse: SelfImprovementReverseClosureObservation{
				Status:             SelfImprovementReverseClosureCompleted,
				ReverseObservation: true,
			},
		},
	)
	if observation.Status != SelfImprovementAdoptionBoundaryAdoptable {
		t.Fatalf("status = %q, want ADOPTABLE", observation.Status)
	}
	if !observation.ContractPreserved || !observation.ReverseObserved {
		t.Fatal("adoptable observation must retain both evidence links")
	}
	if observation.AdoptionAuthorized {
		t.Fatal("adoptable must not authorize adoption")
	}
	if !observation.NonAuthorizing {
		t.Fatal("adoption boundary must remain non-authorizing")
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementAdoptionBoundaryRejectsContractFailure(t *testing.T) {
	observation := ObserveSelfImprovementAdoptionBoundary(
		SelfImprovementAdoptionBoundaryInput{
			CandidateDigest: "candidate-digest",
			Contract: SelfImprovementContractPreservationObservation{
				Status: SelfImprovementContractPreservationRejected,
			},
			Reverse: SelfImprovementReverseClosureObservation{
				Status: SelfImprovementReverseClosureCompleted,
			},
		},
	)
	if observation.Status != SelfImprovementAdoptionBoundaryRejected {
		t.Fatalf("status = %q, want REJECTED", observation.Status)
	}
	if observation.Reason != "CONTRACT_PRESERVATION_REJECTED" {
		t.Fatalf("reason = %q, want contract rejection", observation.Reason)
	}
}

func TestObserveSelfImprovementAdoptionBoundaryKeepsIncompleteEvidenceUnknown(t *testing.T) {
	observation := ObserveSelfImprovementAdoptionBoundary(
		SelfImprovementAdoptionBoundaryInput{
			CandidateDigest: "candidate-digest",
			Contract: SelfImprovementContractPreservationObservation{
				Status: SelfImprovementContractPreservationUnknown,
			},
			Reverse: SelfImprovementReverseClosureObservation{
				Status: SelfImprovementReverseClosureUnknown,
			},
		},
	)
	if observation.Status != SelfImprovementAdoptionBoundaryUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.Reason != "ADOPTION_EVIDENCE_INCOMPLETE" {
		t.Fatalf("reason = %q, want incomplete evidence", observation.Reason)
	}
}