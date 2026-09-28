package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type ChangeProposalStatus string

const (
	ChangeProposalUnknown  ChangeProposalStatus = "UNKNOWN"
	ChangeProposalDeferred ChangeProposalStatus = "DEFERRED"
	ChangeProposalProposed ChangeProposalStatus = "PROPOSED"
	ChangeProposalRejected ChangeProposalStatus = "REJECTED"
)

type ProvenanceChangeProposal struct {
	Status                   ChangeProposalStatus
	SourceDigest             string
	GeneratedArtifactDigest  string
	ReverseObservationDigest string
	TargetScopeDigest        string
	RationaleDigest          string
	ReviewEvidenceDigest     string
	ApplicationBoundaryDigest string
	FirstBoundary            string
	Reason                   string
	ReviewOnly               bool
	ProposalDigest           string
}

// ProposeProvenanceChange turns a verified reverse-observation receipt into a
// review-only proposal. It never applies a mutation or grants authorization.
func ProposeProvenanceChange(receipt ReverseObservationReceipt, requestedChange, targetScopeDigest, rationaleDigest, reviewEvidenceDigest, applicationBoundaryDigest string) ProvenanceChangeProposal {
	proposal := ProvenanceChangeProposal{
		Status:                    ChangeProposalUnknown,
		SourceDigest:              receipt.SourceDigest,
		GeneratedArtifactDigest:   receipt.GeneratedArtifactDigest,
		ReverseObservationDigest:  receipt.ReceiptDigest,
		TargetScopeDigest:         targetScopeDigest,
		RationaleDigest:            rationaleDigest,
		ReviewEvidenceDigest:      reviewEvidenceDigest,
		ApplicationBoundaryDigest: applicationBoundaryDigest,
		FirstBoundary:             "reverse_observation_receipt",
		Reason:                    "proposal inputs are not sufficiently bound",
		ReviewOnly:                true,
	}

	if err := receipt.Validate(); err != nil || receipt.Status == ReverseObservationUnknown {
		proposal.FirstBoundary = "reverse_observation_receipt"
		proposal.Reason = "a valid reverse-observation receipt is required before proposing a change"
	} else if receipt.Status != ReverseObservationObserved {
		proposal.Status = ChangeProposalDeferred
		proposal.FirstBoundary = "reverse_observation"
		proposal.Reason = "incomplete or mismatching reverse observation must remain deferred"
	} else if strings.TrimSpace(requestedChange) == "" {
		proposal.Status = ChangeProposalDeferred
		proposal.FirstBoundary = "requested_change"
		proposal.Reason = "a requested change is required for review"
	} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(requestedChange)), "forbidden:") {
		proposal.Status = ChangeProposalRejected
		proposal.FirstBoundary = "requested_change_scope"
		proposal.Reason = "the requested change declares a forbidden scope"
	} else if !validReverseObservationDigest(targetScopeDigest) {
		proposal.Status = ChangeProposalDeferred
		proposal.FirstBoundary = "target_scope"
		proposal.Reason = "the target scope must be provenance-bound"
	} else if !validReverseObservationDigest(rationaleDigest) {
		proposal.Status = ChangeProposalDeferred
		proposal.FirstBoundary = "rationale"
		proposal.Reason = "a rationale digest is required for review"
	} else if !validReverseObservationDigest(reviewEvidenceDigest) {
		proposal.Status = ChangeProposalDeferred
		proposal.FirstBoundary = "review_evidence"
		proposal.Reason = "review evidence must be provenance-bound"
	} else if !validReverseObservationDigest(applicationBoundaryDigest) {
		proposal.Status = ChangeProposalDeferred
		proposal.FirstBoundary = "application_boundary"
		proposal.Reason = "an explicit application boundary is required"
	} else {
		proposal.Status = ChangeProposalProposed
		proposal.FirstBoundary = "review"
		proposal.Reason = "provenance-bound change proposal is ready for explicit review"
	}
	proposal.ProposalDigest = proposal.digest()
	return proposal
}

func (proposal ProvenanceChangeProposal) Validate() error {
	if !proposal.ReviewOnly {
		return fmt.Errorf("change proposal is not review-only")
	}
	if !validReverseObservationDigest(proposal.ProposalDigest) {
		return fmt.Errorf("change proposal digest is invalid")
	}
	if strings.TrimSpace(proposal.FirstBoundary) == "" || strings.TrimSpace(proposal.Reason) == "" {
		return fmt.Errorf("change proposal boundary or reason is missing")
	}
	switch proposal.Status {
	case ChangeProposalUnknown, ChangeProposalDeferred, ChangeProposalRejected:
	case ChangeProposalProposed:
		for name, value := range map[string]string{
			"source":       proposal.SourceDigest,
			"generated":    proposal.GeneratedArtifactDigest,
			"receipt":      proposal.ReverseObservationDigest,
			"target_scope": proposal.TargetScopeDigest,
			"rationale":    proposal.RationaleDigest,
			"review":       proposal.ReviewEvidenceDigest,
			"boundary":     proposal.ApplicationBoundaryDigest,
		} {
			if !validReverseObservationDigest(value) {
				return fmt.Errorf("proposed change %s digest is invalid", name)
			}
		}
	default:
		return fmt.Errorf("change proposal status %q is invalid", proposal.Status)
	}
	if proposal.digest() != proposal.ProposalDigest {
		return fmt.Errorf("change proposal digest does not match")
	}
	return nil
}

func (proposal ProvenanceChangeProposal) digest() string {
	parts := []string{
		"meta-ontology-provenance-change-proposal",
		string(proposal.Status),
		proposal.SourceDigest,
		proposal.GeneratedArtifactDigest,
		proposal.ReverseObservationDigest,
		proposal.TargetScopeDigest,
		proposal.RationaleDigest,
		proposal.ReviewEvidenceDigest,
		proposal.ApplicationBoundaryDigest,
		proposal.FirstBoundary,
		proposal.Reason,
		fmt.Sprintf("%t", proposal.ReviewOnly),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(sum[:])
}
