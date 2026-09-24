package lsp

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveSelfImprovementAdoptionBoundaryLSPProjectsEvidence(t *testing.T) {
	observation := ObserveSelfImprovementAdoptionBoundaryLSP(
		"file:///workspace/main.gooo",
		12,
		"Improve",
		valueexecution.SelfImprovementAdoptionBoundaryObservation{
			Status:             valueexecution.SelfImprovementAdoptionBoundaryAdoptable,
			CandidateDigest:    "candidate-digest",
			ContractStatus:     valueexecution.SelfImprovementContractPreservationAccepted,
			ReverseStatus:      valueexecution.SelfImprovementReverseClosureCompleted,
			ContractPreserved:  true,
			ReverseObserved:    true,
			Reason:             "EVIDENCE_COMPLETE_ADOPTION_PENDING",
			Digest:             "origin-digest",
		},
	)
	if !observation.Visible {
		t.Fatal("complete adoption evidence should be visible")
	}
	if observation.Status != valueexecution.SelfImprovementAdoptionBoundaryAdoptable {
		t.Fatalf("status = %q, want ADOPTABLE", observation.Status)
	}
	if !observation.ContractPreserved || !observation.ReverseObserved {
		t.Fatal("LSP projection must retain evidence links")
	}
	if observation.AdoptionAuthorized {
		t.Fatal("LSP projection must not authorize adoption")
	}
	if !observation.NonAuthorizing {
		t.Fatal("LSP projection must remain non-authorizing")
	}
	if observation.OriginDigest != "origin-digest" {
		t.Fatalf("origin digest = %q, want origin-digest", observation.OriginDigest)
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementAdoptionBoundaryLSPRejectsInvalidContext(t *testing.T) {
	observation := ObserveSelfImprovementAdoptionBoundaryLSP(
		"",
		12,
		"Improve",
		valueexecution.SelfImprovementAdoptionBoundaryObservation{
			Status:          valueexecution.SelfImprovementAdoptionBoundaryRejected,
			CandidateDigest: "candidate-digest",
			Digest:          "origin-digest",
		},
	)
	if observation.Visible {
		t.Fatal("invalid LSP context should not be visible")
	}
	if observation.Status != valueexecution.SelfImprovementAdoptionBoundaryUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.Reason != "LSP_CONTEXT_INVALID" {
		t.Fatalf("reason = %q, want LSP_CONTEXT_INVALID", observation.Reason)
	}
}

func TestObserveSelfImprovementAdoptionBoundaryLSPHidesUnknownSource(t *testing.T) {
	observation := ObserveSelfImprovementAdoptionBoundaryLSP(
		"file:///workspace/main.gooo",
		12,
		"Improve",
		valueexecution.SelfImprovementAdoptionBoundaryObservation{
			Status: valueexecution.SelfImprovementAdoptionBoundaryUnknown,
		},
	)
	if observation.Visible {
		t.Fatal("unknown source should not be visible")
	}
	if observation.Status != valueexecution.SelfImprovementAdoptionBoundaryUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.Reason != "SOURCE_OBSERVATION_UNKNOWN" {
		t.Fatalf("reason = %q, want SOURCE_OBSERVATION_UNKNOWN", observation.Reason)
	}
}