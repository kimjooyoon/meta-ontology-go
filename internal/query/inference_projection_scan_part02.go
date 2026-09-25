package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func (projection InferenceProjection) scanEvidence(
	request InferenceQuery, budget *inferenceWorkBudget, edges []InferenceRow,
	claims []SemanticChangeRow, selectedEvidence map[semantic.ID]struct{},
) ([]InferenceEvidenceRow, error) {
	evidence := make([]InferenceEvidenceRow, 0)
	if !request.IncludeEvidence && request.EvidenceID == "" {
		return evidence, nil
	}
	for _, record := range projection.path.Evidence {
		if !budget.evidence() {
			return nil, ErrInferenceQueryBudget
		}
		if request.EvidenceID != "" && ID(record.ID.String()) != request.EvidenceID {
			continue
		}
		if !request.IncludeEvidence {
			if _, selected := selectedEvidence[record.ID]; !selected {
				continue
			}
		}
		if request.IncludeEvidence && request.EvidenceID == "" && request.hasSelectors() {
			if _, selected := selectedEvidence[record.ID]; !selected {
				continue
			}
		}
		if len(edges)+len(claims)+len(evidence) >= request.Limit {
			return nil, ErrInferenceQueryBudget
		}
		evidence = append(evidence, inferenceEvidenceRowFromRecord(record))
	}
	return evidence, nil
}
