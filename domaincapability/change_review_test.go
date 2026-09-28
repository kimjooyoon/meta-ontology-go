package domaincapability

import "testing"

func TestReviewProvenanceChangeAcceptsOnlyExplicitApplicationBoundary(t *testing.T) {
	proposal := ProposeProvenanceChange(proposalTestReceipt(), "add reverse observation", proposalTestDigest("5"), proposalTestDigest("6"), proposalTestDigest("7"), proposalTestDigest("8"))
	review := ReviewProvenanceChange(proposal, proposalTestDigest("9"), proposalTestDigest("a"), proposalTestDigest("b"))
	if review.Status != ChangeReviewAccepted || !review.ApplicationAllowed || review.FirstBoundary != "application_boundary" {
		t.Fatalf("unexpected accepted review: %+v", review)
	}
	if err := review.Validate(); err != nil {
		t.Fatalf("accepted review should validate: %v", err)
	}
}

func TestReviewProvenanceChangePreservesMissingApproval(t *testing.T) {
	proposal := ProposeProvenanceChange(proposalTestReceipt(), "add reverse observation", proposalTestDigest("5"), proposalTestDigest("6"), proposalTestDigest("7"), proposalTestDigest("8"))
	review := ReviewProvenanceChange(proposal, proposalTestDigest("9"), "", proposalTestDigest("b"))
	if review.Status != ChangeReviewPending || review.FirstBoundary != "approval_evidence" || review.ApplicationAllowed {
		t.Fatalf("unexpected pending review: %+v", review)
	}
	if err := review.Validate(); err != nil {
		t.Fatalf("pending review should validate: %v", err)
	}
}
