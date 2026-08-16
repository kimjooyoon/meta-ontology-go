package coupling

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type inferenceIndex struct {
	edges    map[semantic.ID]semantic.InferenceEdge
	claims   map[semantic.ID]semantic.SemanticChangeClaim
	evidence map[semantic.ID]semantic.InferenceEvidence
}

func semanticKindForClaim(claim ChangeClaim) (semantic.SemanticChangeKind, bool) {
	switch claim {
	case ChangeClaimDelta:
		return semantic.SemanticDelta, true
	case ChangeClaimNoDelta:
		return semantic.NoSemanticDelta, true
	default:
		return "", false
	}
}

func normalizeInferencePath(path semantic.InferencePathV1) (semantic.InferencePathV1, *evaluationIssue) {
	if path.Version == "" && path.Edges == nil && path.Claims == nil && path.Evidence == nil {
		return semantic.InferencePathV1{}, required("inference path")
	}
	normalized, err := path.Normalized()
	if err != nil {
		return semantic.InferencePathV1{}, failIssue(ReasonInferencePathMalformed, "inference path")
	}
	return normalized, nil
}

func indexInferencePath(path semantic.InferencePathV1) inferenceIndex {
	index := inferenceIndex{
		edges:    make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges)),
		claims:   make(map[semantic.ID]semantic.SemanticChangeClaim, len(path.Claims)),
		evidence: make(map[semantic.ID]semantic.InferenceEvidence, len(path.Evidence)),
	}
	for _, edge := range path.Edges {
		index.edges[edge.RecordID] = edge
	}
	for _, claim := range path.Claims {
		index.claims[claim.RecordID] = claim
	}
	for _, evidence := range path.Evidence {
		index.evidence[evidence.ID] = evidence
	}
	return index
}

func recordMentionsOwner(record semantic.InferenceRecord, owner semantic.ID) bool {
	return record.SubjectID == owner || record.ObjectID == owner
}

func validateReceiptPath(
	receipt CouplingReceipt, entry ManifestEntry, path semantic.InferencePathV1,
) *evaluationIssue {
	index := indexInferencePath(path)
	claim, exists := index.claims[receipt.InferenceClaimID]
	if !exists {
		return failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
	}
	wantKind, validClaim := semanticKindForClaim(receipt.ChangeClaim)
	if !validClaim || receipt.ReceiptKind != wantKind {
		return failIssue(ReasonContradictoryReceipt, receipt.SurfaceID.String())
	}
	if !recordMentionsOwner(claim.InferenceRecord, receipt.SemanticOwnerID) ||
		claim.Kind != wantKind || claim.Before.Semantic != receipt.BeforeCanonicalSemanticDigest ||
		claim.After.Semantic != receipt.AfterCanonicalSemanticDigest ||
		claim.Before.Source != receipt.BeforeAuthoritySourceDigest || claim.After.Source != receipt.AfterAuthoritySourceDigest {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if claim.Kind == semantic.SemanticDelta &&
		(claim.CanonicalDelta != receipt.CanonicalDelta || claim.DeltaDigest != receipt.DeltaDigest) {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	if claim.Kind == semantic.NoSemanticDelta && (receipt.CanonicalDelta != "" || receipt.DeltaDigest != "") {
		return failIssue(ReasonNoDeltaWithoutEquality, receipt.SurfaceID.String())
	}
	if len(receipt.OriginPathIDs) == 0 {
		return failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
	}
	selected := make([]semantic.InferenceEdge, 0, len(receipt.OriginPathIDs))
	seenPathIDs := make(map[semantic.ID]struct{}, len(receipt.OriginPathIDs))
	for _, pathID := range receipt.OriginPathIDs {
		if _, duplicate := seenPathIDs[pathID]; duplicate {
			return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		seenPathIDs[pathID] = struct{}{}
		edge, ok := index.edges[pathID]
		if !ok {
			return failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
		}
		selected = append(selected, edge)
	}
	chain, err := semantic.NewInferencePathChain(selected...)
	if err != nil {
		return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
	}
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
	if receipt.ChangeClaim == ChangeClaimDelta {
		if receipt.AuthoritativeSource == nil || receipt.AuthoritativeSource.SourceID == "" || receipt.AuthoritativeSource.Path == "" {
			return failIssue(ReasonDeltaWithoutSource, receipt.SurfaceID.String())
		}
		rootFound := false
		for _, edge := range chain.Edges {
			for _, root := range edge.SourceRoots {
				if root == receipt.AuthoritativeSource.SourceID {
					rootFound = true
				}
			}
		}
		if !rootFound {
			return failIssue(ReasonMissingAuthorityPath, receipt.SurfaceID.String())
		}
	}
	return validateEvidenceReferences(receipt, index, chain.Edges, claim)
}

func validateEvidenceReferences(
	receipt CouplingReceipt, index inferenceIndex, edges []semantic.InferenceEdge, claim semantic.SemanticChangeClaim,
) *evaluationIssue {
	expected := make(map[semantic.ID]semantic.EvidenceReference)
	add := func(ref semantic.EvidenceReference) *evaluationIssue {
		if previous, exists := expected[ref.ID]; exists && previous.Digest != ref.Digest {
			return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
		}
		expected[ref.ID] = ref
		return nil
	}
	for _, ref := range claim.Evidence {
		if issue := add(ref); issue != nil {
			return issue
		}
	}
	for _, edge := range edges {
		for _, ref := range edge.Evidence {
			if issue := add(ref); issue != nil {
				return issue
			}
		}
	}
	actual := append([]semantic.EvidenceReference(nil), receipt.EvidenceRefs...)
	sort.Slice(actual, func(i, j int) bool { return actual[i].ID < actual[j].ID })
	if len(actual) != len(expected) {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	independent := false
	for i, ref := range actual {
		if i > 0 && ref.ID == actual[i-1].ID {
			return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		want, exists := expected[ref.ID]
		if !exists || want.Digest != ref.Digest {
			return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
		}
		evidence, exists := index.evidence[ref.ID]
		if !exists || evidence.Digest != ref.Digest {
			return failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
		}
		if evidence.Independent {
			independent = true
		}
	}
	if !independent || !terminalHasIndependentEvidence(edges[len(edges)-1], index) {
		return failIssue(ReasonMissingVerification, receipt.SurfaceID.String())
	}
	return nil
}

func terminalHasIndependentEvidence(edge semantic.InferenceEdge, index inferenceIndex) bool {
	for _, ref := range edge.Evidence {
		if evidence, ok := index.evidence[ref.ID]; ok && evidence.Independent {
			return true
		}
	}
	return false
}
