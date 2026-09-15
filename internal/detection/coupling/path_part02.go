package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func receiptClaim(receipt CouplingReceipt, index inferenceIndex) (semantic.SemanticChangeClaim, *evaluationIssue) {
	claim, exists := index.claims[receipt.InferenceClaimID]
	if !exists {
		return semantic.SemanticChangeClaim{}, failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
	}
	wantKind, validClaim := semanticKindForClaim(receipt.ChangeClaim)
	if !validClaim || receipt.ReceiptKind != wantKind {
		return semantic.SemanticChangeClaim{}, failIssue(ReasonContradictoryReceipt, receipt.SurfaceID.String())
	}
	if !recordMentionsOwner(claim.InferenceRecord, receipt.SemanticOwnerID) ||
		claim.Kind != wantKind || claim.Before.Semantic != receipt.BeforeCanonicalSemanticDigest ||
		claim.After.Semantic != receipt.AfterCanonicalSemanticDigest ||
		claim.Before.Source != receipt.BeforeAuthoritySourceDigest || claim.After.Source != receipt.AfterAuthoritySourceDigest {
		return semantic.SemanticChangeClaim{}, failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if claim.Kind == semantic.SemanticDelta &&
		(claim.CanonicalDelta != receipt.CanonicalDelta || claim.DeltaDigest != receipt.DeltaDigest) {
		return semantic.SemanticChangeClaim{}, failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if claim.Kind == semantic.NoSemanticDelta && (receipt.CanonicalDelta != "" || receipt.DeltaDigest != "") {
		return semantic.SemanticChangeClaim{}, failIssue(ReasonNoDeltaWithoutEquality, receipt.SurfaceID.String())
	}
	return claim, nil
}
func selectedReceiptChain(receipt CouplingReceipt, index inferenceIndex) (semantic.InferencePathChain, *evaluationIssue) {
	if len(receipt.OriginPathIDs) == 0 {
		return semantic.InferencePathChain{}, failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
	}
	selected := make([]semantic.InferenceEdge, 0, len(receipt.OriginPathIDs))
	seenPathIDs := make(map[semantic.ID]struct{}, len(receipt.OriginPathIDs))
	for _, pathID := range receipt.OriginPathIDs {
		if _, duplicate := seenPathIDs[pathID]; duplicate {
			return semantic.InferencePathChain{}, failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		seenPathIDs[pathID] = struct{}{}
		edge, ok := index.edges[pathID]
		if !ok {
			return semantic.InferencePathChain{}, failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
		}
		selected = append(selected, edge)
	}
	chain, err := semantic.NewInferencePathChain(selected...)
	if err != nil {
		return semantic.InferencePathChain{}, failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
	}
	return chain, nil
}
