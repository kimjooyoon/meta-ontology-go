package coupling

import (
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
	claim, issue := receiptClaim(receipt, index)
	if issue != nil {
		return issue
	}
	chain, issue := selectedReceiptChain(receipt, index)
	if issue != nil {
		return issue
	}
	if issue := validateReceiptChain(receipt, entry, index, chain); issue != nil {
		return issue
	}
	return validateEvidenceReferences(receipt, index, chain.Edges, claim)
}
