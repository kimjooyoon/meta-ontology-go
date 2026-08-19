package query

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func inferenceRowFromEdge(edge semantic.InferenceEdge) InferenceRow {
	record := edge.InferenceRecord
	roots := make([]ID, 0, len(edge.SourceRoots))
	for _, root := range edge.SourceRoots {
		roots = append(roots, ID(root.String()))
	}
	return InferenceRow{
		RecordID: ID(record.RecordID.String()), SubjectID: ID(record.SubjectID.String()), ObjectID: ID(record.ObjectID.String()),
		Kind: edge.Kind, Phase: record.Phase.Phase, PhaseOrdinal: record.Phase.Ordinal,
		AuthorityLayer: record.Authority.Layer, AuthorityEffect: record.Authority.Effect,
		Rule: record.Rule, Before: record.Before, After: record.After,
		Evidence:    append([]semantic.EvidenceReference(nil), record.Evidence...),
		SourceRoots: roots, AcceptanceReceipt: ID(edge.AcceptanceReceipt.String()),
	}
}
func semanticChangeRowFromClaim(claim semantic.SemanticChangeClaim) SemanticChangeRow {
	record := claim.InferenceRecord
	return SemanticChangeRow{
		RecordID: ID(record.RecordID.String()), SubjectID: ID(record.SubjectID.String()), ObjectID: ID(record.ObjectID.String()),
		Kind: claim.Kind, Phase: record.Phase.Phase, PhaseOrdinal: record.Phase.Ordinal,
		AuthorityLayer: record.Authority.Layer, AuthorityEffect: record.Authority.Effect,
		Rule: record.Rule, Before: record.Before, After: record.After,
		Evidence:       append([]semantic.EvidenceReference(nil), record.Evidence...),
		CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest,
	}
}
func inferenceEvidenceRowFromRecord(record semantic.InferenceEvidence) InferenceEvidenceRow {
	return InferenceEvidenceRow{
		ID: ID(record.ID.String()), Digest: record.Digest, Before: record.Before, After: record.After,
		SourceBacked: record.SourceBacked, Independent: record.Independent, Controls: record.Controls,
	}
}
func (projection InferenceProjection) budgetFailure(request InferenceQuery, requestHash string, budget *inferenceWorkBudget) (InferenceQueryResult, error) {
	return rejectedInferenceResponseWithWork(request, requestHash, budget.work, fmt.Errorf("%w: limit %d", ErrInferenceQueryBudget, request.MaxWork))
}
func (projection InferenceProjection) limitFailure(request InferenceQuery, requestHash string, budget *inferenceWorkBudget) (InferenceQueryResult, error) {
	return rejectedInferenceResponseWithWork(request, requestHash, budget.work, fmt.Errorf("%w: maximum rows %d", ErrInferenceQueryBudget, request.Limit))
}
func (projection InferenceProjection) staleFailure(request InferenceQuery, requestHash string, budget *inferenceWorkBudget) (InferenceQueryResult, error) {
	return rejectedInferenceResponseWithWork(request, requestHash, budget.work, ErrInferenceStaleSnapshot)
}
func (projection InferenceProjection) chainFailure(request InferenceQuery, requestHash string, budget *inferenceWorkBudget, err error) (InferenceQueryResult, error) {
	return rejectedInferenceResponseWithWork(request, requestHash, budget.work, err)
}
func rejectedInferenceResponse(request InferenceQuery, err error) (InferenceQueryResult, error) {
	normalized, normalizeErr := request.normalized()
	if normalizeErr == nil {
		request = normalized
	}
	requestHash, _ := request.CanonicalDigest()
	workLimit := request.MaxWork
	if workLimit < 0 {
		workLimit = 0
	}
	return rejectedInferenceResponseWithWork(request, requestHash, InferenceWork{Limit: workLimit}, err)
}
