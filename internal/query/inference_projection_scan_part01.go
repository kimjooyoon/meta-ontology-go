package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func (projection InferenceProjection) scanEdges(
	request InferenceQuery, budget *inferenceWorkBudget,
) ([]InferenceRow, []semantic.InferenceEdge, map[semantic.ID]struct{}, error) {
	rows := make([]InferenceRow, 0)
	matched := make([]semantic.InferenceEdge, 0)
	selectedEvidence := make(map[semantic.ID]struct{})
	for _, edge := range projection.path.Edges {
		if !budget.edge() {
			return nil, nil, nil, ErrInferenceQueryBudget
		}
		if inferenceRecordStale(request, edge.InferenceRecord) &&
			inferenceRecordIdentityMatches(request, edge.InferenceRecord, edge.Kind, "", false) &&
			evidenceReferencesMatch(request.EvidenceID, edge.Evidence) {
			return nil, nil, nil, ErrInferenceStaleSnapshot
		}
		if !inferenceRecordMatches(request, edge.InferenceRecord, edge.Kind, "", false) ||
			!evidenceReferencesMatch(request.EvidenceID, edge.Evidence) {
			continue
		}
		if len(rows) >= request.Limit {
			return nil, nil, nil, ErrInferenceQueryBudget
		}
		rows = append(rows, inferenceRowFromEdge(edge))
		matched = append(matched, edge)
		for _, ref := range edge.Evidence {
			selectedEvidence[ref.ID] = struct{}{}
		}
	}
	return rows, matched, selectedEvidence, nil
}
func (projection InferenceProjection) scanClaims(
	request InferenceQuery, budget *inferenceWorkBudget, edges []InferenceRow,
	selectedEvidence map[semantic.ID]struct{},
) ([]SemanticChangeRow, error) {
	claims := make([]SemanticChangeRow, 0)
	if !request.IncludeClaims {
		return claims, nil
	}
	for _, claim := range projection.path.Claims {
		if !budget.claim() {
			return nil, ErrInferenceQueryBudget
		}
		if inferenceRecordStale(request, claim.InferenceRecord) &&
			inferenceRecordIdentityMatches(request, claim.InferenceRecord, "", claim.Kind, true) &&
			evidenceReferencesMatch(request.EvidenceID, claim.Evidence) {
			return nil, ErrInferenceStaleSnapshot
		}
		if !inferenceRecordMatches(request, claim.InferenceRecord, "", claim.Kind, true) ||
			!evidenceReferencesMatch(request.EvidenceID, claim.Evidence) {
			continue
		}
		if len(edges)+len(claims) >= request.Limit {
			return nil, ErrInferenceQueryBudget
		}
		claims = append(claims, semanticChangeRowFromClaim(claim))
		for _, ref := range claim.Evidence {
			selectedEvidence[ref.ID] = struct{}{}
		}
	}
	return claims, nil
}
