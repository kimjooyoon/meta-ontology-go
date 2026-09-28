package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type ChangeReviewStatus string

const (
	ChangeReviewUnknown   ChangeReviewStatus = "UNKNOWN"
	ChangeReviewPending   ChangeReviewStatus = "PENDING"
	ChangeReviewAccepted  ChangeReviewStatus = "ACCEPTED_FOR_APPLICATION"
	ChangeReviewRejected  ChangeReviewStatus = "REJECTED"
)

type ProvenanceChangeReview struct {
	Status             ChangeReviewStatus
	ProposalDigest     string
	ReviewerDigest     string
	ApprovalEvidenceDigest string
	ApplicationBoundaryDigest string
	FirstBoundary      string
	Reason             string
	ApplicationAllowed bool
	ReviewDigest       string
}

// ReviewProvenanceChange approves only the evidence state needed by a later
// explicit application boundary. It never applies a proposal itself.
func ReviewProvenanceChange(proposal ProvenanceChangeProposal, reviewerDigest, approvalEvidenceDigest, applicationBoundaryDigest string) ProvenanceChangeReview {
	review := ProvenanceChangeReview{
		Status:                    ChangeReviewUnknown,
		ProposalDigest:            proposal.ProposalDigest,
		ReviewerDigest:            reviewerDigest,
		ApprovalEvidenceDigest:    approvalEvidenceDigest,
		ApplicationBoundaryDigest: applicationBoundaryDigest,
		FirstBoundary:             "proposal",
		Reason:                    "a valid proposed change is required before review",
		ApplicationAllowed:        false,
	}
	if err := proposal.Validate(); err != nil || proposal.Status != ChangeProposalProposed {
		review.Status = ChangeReviewPending
		review.FirstBoundary = "proposal"
		review.Reason = "proposal evidence is incomplete or not yet proposed"
	} else if !validReverseObservationDigest(reviewerDigest) {
		review.Status = ChangeReviewPending
		review.FirstBoundary = "reviewer"
		review.Reason = "reviewer identity evidence is required"
	} else if !validReverseObservationDigest(approvalEvidenceDigest) {
		review.Status = ChangeReviewPending
		review.FirstBoundary = "approval_evidence"
		review.Reason = "approval evidence is required"
	} else if !validReverseObservationDigest(applicationBoundaryDigest) {
		review.Status = ChangeReviewPending
		review.FirstBoundary = "application_boundary"
		review.Reason = "an explicit application boundary is required"
	} else {
		review.Status = ChangeReviewAccepted
		review.FirstBoundary = "application_boundary"
		review.Reason = "proposal is accepted for a separate explicit application step"
		review.ApplicationAllowed = true
	}
	review.ReviewDigest = review.digest()
	return review
}

func (review ProvenanceChangeReview) Validate() error {
	if !validReverseObservationDigest(review.ReviewDigest) {
		return fmt.Errorf("change review digest is invalid")
	}
	if review.Status == ChangeReviewAccepted && !review.ApplicationAllowed {
		return fmt.Errorf("accepted change review does not allow its declared boundary")
	}
	if review.Status != ChangeReviewAccepted && review.ApplicationAllowed {
		return fmt.Errorf("non-accepted change review allows application")
	}
	if strings.TrimSpace(review.FirstBoundary) == "" || strings.TrimSpace(review.Reason) == "" {
		return fmt.Errorf("change review boundary or reason is missing")
	}
	switch review.Status {
	case ChangeReviewUnknown, ChangeReviewPending, ChangeReviewRejected:
	case ChangeReviewAccepted:
		for name, value := range map[string]string{
			"proposal":    review.ProposalDigest,
			"reviewer":    review.ReviewerDigest,
			"approval":    review.ApprovalEvidenceDigest,
			"boundary":    review.ApplicationBoundaryDigest,
		} {
			if !validReverseObservationDigest(value) {
				return fmt.Errorf("accepted change review %s digest is invalid", name)
			}
		}
	default:
		return fmt.Errorf("change review status %q is invalid", review.Status)
	}
	if review.digest() != review.ReviewDigest {
		return fmt.Errorf("change review digest does not match")
	}
	return nil
}

func (review ProvenanceChangeReview) digest() string {
	parts := []string{
		"meta-ontology-provenance-change-review",
		string(review.Status),
		review.ProposalDigest,
		review.ReviewerDigest,
		review.ApprovalEvidenceDigest,
		review.ApplicationBoundaryDigest,
		review.FirstBoundary,
		review.Reason,
		fmt.Sprintf("%t", review.ApplicationAllowed),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(sum[:])
}
