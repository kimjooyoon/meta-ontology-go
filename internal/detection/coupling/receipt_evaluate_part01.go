package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateReceiptReferences(receipt CouplingReceipt) *evaluationIssue {
	if _, issue := normalizeID(receipt.InferenceClaimID, "inference claim ID"); issue != nil {
		return issue
	}
	seen := make(map[semantic.ID]struct{}, len(receipt.OriginPathIDs))
	for _, pathID := range receipt.OriginPathIDs {
		if _, issue := normalizeID(pathID, "origin path ID"); issue != nil {
			return issue
		}
		if _, duplicate := seen[pathID]; duplicate {
			return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		seen[pathID] = struct{}{}
	}
	seenEvidence := make(map[semantic.ID]struct{}, len(receipt.EvidenceRefs))
	for _, ref := range receipt.EvidenceRefs {
		if _, issue := normalizeID(ref.ID, "evidence ID"); issue != nil {
			return issue
		}
		if issue := normalizeDigestValue(ref.Digest, "evidence digest"); issue != nil {
			return issue
		}
		if _, duplicate := seenEvidence[ref.ID]; duplicate {
			return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		seenEvidence[ref.ID] = struct{}{}
	}
	return nil
}
