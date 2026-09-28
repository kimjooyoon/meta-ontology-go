package domaincapability

import (
	"strings"
	"testing"
)

func proposalTestDigest(char string) string {
	return "sha256:" + strings.Repeat(char, 64)
}

func proposalTestReceipt() ReverseObservationReceipt {
	input := ReverseObservationInput{
		SourceDigest:            proposalTestDigest("0"),
		DeclarationDigest:       proposalTestDigest("1"),
		IRDigest:                proposalTestDigest("2"),
		GeneratedArtifactDigest: proposalTestDigest("3"),
		ObservedArtifactDigest:  proposalTestDigest("3"),
		EvidencePrefixDigest:    proposalTestDigest("4"),
	}
	return ObserveReverseObservation(input)
}

func TestProposeProvenanceChangeRequiresExplicitReviewInputs(t *testing.T) {
	proposal := ProposeProvenanceChange(proposalTestReceipt(), "add reverse observation", proposalTestDigest("5"), proposalTestDigest("6"), proposalTestDigest("7"), proposalTestDigest("8"))
	if proposal.Status != ChangeProposalProposed || !proposal.ReviewOnly || proposal.FirstBoundary != "review" {
		t.Fatalf("unexpected proposed change: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("proposed change should validate: %v", err)
	}
}

func TestProposeProvenanceChangeDefersMismatch(t *testing.T) {
	input := ReverseObservationInput{
		SourceDigest:            proposalTestDigest("0"),
		DeclarationDigest:       proposalTestDigest("1"),
		IRDigest:                proposalTestDigest("2"),
		GeneratedArtifactDigest: proposalTestDigest("3"),
		ObservedArtifactDigest:  proposalTestDigest("9"),
		EvidencePrefixDigest:    proposalTestDigest("4"),
	}
	proposal := ProposeProvenanceChange(ObserveReverseObservation(input), "add reverse observation", proposalTestDigest("5"), proposalTestDigest("6"), proposalTestDigest("7"), proposalTestDigest("8"))
	if proposal.Status != ChangeProposalDeferred || proposal.FirstBoundary != "reverse_observation" {
		t.Fatalf("unexpected deferred change: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("deferred change should validate: %v", err)
	}
}

func TestProposeProvenanceChangeRejectsForbiddenScope(t *testing.T) {
	proposal := ProposeProvenanceChange(proposalTestReceipt(), "forbidden: mutate catalog", proposalTestDigest("5"), proposalTestDigest("6"), proposalTestDigest("7"), proposalTestDigest("8"))
	if proposal.Status != ChangeProposalRejected || proposal.FirstBoundary != "requested_change_scope" {
		t.Fatalf("unexpected rejected change: %+v", proposal)
	}
	if err := proposal.Validate(); err != nil {
		t.Fatalf("rejected change should validate: %v", err)
	}
}
