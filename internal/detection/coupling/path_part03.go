package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateReceiptChain(
	receipt CouplingReceipt, entry ManifestEntry, index inferenceIndex, chain semantic.InferencePathChain,
) *evaluationIssue {
	if len(chain.Edges) == 0 {
		return failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
	}
	first := chain.Edges[0]
	if first.Kind != semantic.InferenceAuthoritativeDeclaration && first.Kind != semantic.InferenceAcceptedLift {
		if first.Kind == semantic.InferenceObservationCandidate {
			return failIssue(ReasonCandidateOnlyPath, receipt.SurfaceID.String())
		}
		return failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
	}
	if first.SubjectID != receipt.SemanticOwnerID || first.ObjectID != receipt.CodeSymbolID || len(first.SourceRoots) == 0 {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if first.Before.Source != receipt.BeforeAuthoritySourceDigest || first.After.Source != receipt.AfterAuthoritySourceDigest ||
		first.Before.Semantic != receipt.BeforeCanonicalSemanticDigest || first.After.Semantic != receipt.AfterCanonicalSemanticDigest {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if first.Kind == semantic.InferenceAcceptedLift && first.AcceptanceReceipt == "" {
		return failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
	}
	last := chain.Edges[len(chain.Edges)-1]
	if last.Kind != semantic.InferenceIndependentVerification || last.SubjectID != entry.SurfaceID ||
		last.Before.Semantic != receipt.BeforeCanonicalSemanticDigest || last.After.Semantic != receipt.AfterCanonicalSemanticDigest {
		return failIssue(ReasonMissingVerification, receipt.SurfaceID.String())
	}
	if _, exists := index.evidence[last.ObjectID]; !exists {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	foundCode, foundSurface, projection := false, false, false
	for _, edge := range chain.Edges {
		if edge.SubjectID == receipt.CodeSymbolID || edge.ObjectID == receipt.CodeSymbolID {
			foundCode = true
		}
		if edge.SubjectID == entry.SurfaceID || edge.ObjectID == entry.SurfaceID {
			foundSurface = true
		}
		if edge.Kind == semantic.InferenceDerivedProjection && edge.SubjectID == receipt.CodeSymbolID && edge.ObjectID == entry.SurfaceID {
			projection = true
		}
	}
	if !foundCode || !foundSurface || !projection {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if receipt.ChangeClaim != ChangeClaimDelta {
		return nil
	}
	if receipt.AuthoritativeSource == nil || receipt.AuthoritativeSource.SourceID == "" || receipt.AuthoritativeSource.Path == "" {
		return failIssue(ReasonDeltaWithoutSource, receipt.SurfaceID.String())
	}
	for _, edge := range chain.Edges {
		for _, root := range edge.SourceRoots {
			if root == receipt.AuthoritativeSource.SourceID {
				return nil
			}
		}
	}
	return failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
}
