package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

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
