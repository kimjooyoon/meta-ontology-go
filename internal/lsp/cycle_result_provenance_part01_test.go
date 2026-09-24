package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveSelfImprovementCycleResultProvenanceKeepsStatusVisible(t *testing.T) {
	observation := ObserveSelfImprovementCycleResultProvenance(
		DeclarationCompletionContext{
			DocumentURI:       "file:///workspace/examples/self-improvement/main.gooo",
			DocumentVersion:   18,
			DeclarationSymbol: "activity Observe",
			EnvironmentDigest: "environment-digest",
		},
		provenance.OriginChainObservation{OriginTransition: provenance.OriginChainTransitionChanged},
		valueexecution.SelfImprovementCycleResultObservation{
			CycleDigest:        "cycle-digest",
			OutcomeDigest:      "outcome-digest",
			Status:             valueexecution.SelfImprovementCycleResultStatusCompleted,
			NonAuthorizing:     true,
			ReverseObservation: true,
			EvidenceCount:      4,
			Digest:             lspSelfImprovementDigest("cycle-result"),
		},
	)
	if observation.Decision != SelfImprovementCycleResultProvenanceVisible || observation.Reason != "CYCLE_RESULT_STATUS_BOUND" {
		t.Fatalf("observation=%#v, want visible cycle result", observation)
	}
	if observation.EvidenceCount != 4 || !observation.NonAuthorizing {
		t.Fatalf("observation=%#v, want evidence count and non-authorizing boundary", observation)
	}
	if err := ValidateSelfImprovementCycleResultProvenance(observation); err != nil {
		t.Fatalf("visible cycle result should validate: %v", err)
	}
}

func TestObserveSelfImprovementCycleResultProvenancePreservesUnknown(t *testing.T) {
	observation := ObserveSelfImprovementCycleResultProvenance(
		DeclarationCompletionContext{
			DocumentURI:       "file:///workspace/examples/self-improvement/main.gooo",
			DocumentVersion:   18,
			DeclarationSymbol: "activity Observe",
			EnvironmentDigest: "environment-digest",
		},
		provenance.OriginChainObservation{OriginTransition: provenance.OriginChainTransitionChanged},
		valueexecution.SelfImprovementCycleResultObservation{
			Status:          valueexecution.SelfImprovementCycleResultStatusUnknown,
			NonAuthorizing:  true,
			ReverseObservation: true,
			Digest:          lspSelfImprovementDigest("unknown-cycle-result"),
		},
	)
	if observation.Decision != SelfImprovementCycleResultProvenanceUnknown || observation.Reason != "CYCLE_RESULT_UNKNOWN" {
		t.Fatalf("observation=%#v, want UNKNOWN cycle result", observation)
	}
	if err := ValidateSelfImprovementCycleResultProvenance(observation); err != nil {
		t.Fatalf("unknown cycle result should validate: %v", err)
	}
}
