package query

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func (projection InferenceProjection) explanationChain(
	request InferenceQuery, matched []semantic.InferenceEdge, rows []InferenceRow, budget *inferenceWorkBudget,
) (InferenceChainResult, error) {
	if len(matched) == 0 {
		return InferenceChainResult{}, fmt.Errorf("%w: path_orphan: no matching edges", ErrInferenceChain)
	}
	if len(matched) > request.MaxDepth {
		return InferenceChainResult{}, fmt.Errorf("%w: depth limit %d exceeded", ErrInferenceQueryBudget, request.MaxDepth)
	}
	for range matched {
		if !budget.chain() {
			return InferenceChainResult{}, ErrInferenceQueryBudget
		}
	}
	chain, err := semantic.NewInferencePathChain(matched...)
	if err != nil {
		return InferenceChainResult{}, fmt.Errorf("%w: %v", ErrInferenceChain, err)
	}
	if request.ChainStartID != "" && ID(chain.Edges[0].SubjectID.String()) != request.ChainStartID {
		return InferenceChainResult{}, fmt.Errorf("%w: path_orphan: chain start does not match request", ErrInferenceChain)
	}
	if request.ChainEndID != "" && ID(chain.Edges[len(chain.Edges)-1].ObjectID.String()) != request.ChainEndID {
		return InferenceChainResult{}, fmt.Errorf("%w: path_orphan: chain end does not match request", ErrInferenceChain)
	}
	byID := make(map[ID]InferenceRow, len(rows))
	for _, row := range rows {
		byID[row.RecordID] = row
	}
	ordered := make([]InferenceRow, 0, len(chain.Edges))
	for _, edge := range chain.Edges {
		row, ok := byID[ID(edge.RecordID.String())]
		if !ok {
			return InferenceChainResult{}, fmt.Errorf("%w: path_orphan: chain edge was not selected", ErrInferenceChain)
		}
		ordered = append(ordered, row)
	}
	return InferenceChainResult{Chain: chain, Edges: ordered, Depth: len(ordered), Complete: true}, nil
}
func inferenceRecordIdentityMatches(
	query InferenceQuery, record semantic.InferenceRecord, kind semantic.InferenceKind,
	claimKind semantic.SemanticChangeKind, isClaim bool,
) bool {
	copyQuery := query
	copyQuery.Before = semantic.SnapshotDigests{}
	copyQuery.After = semantic.SnapshotDigests{}
	copyQuery.Controls = semantic.InferenceControls{}
	return inferenceRecordMatches(copyQuery, record, kind, claimKind, isClaim)
}
func inferenceRecordStale(query InferenceQuery, record semantic.InferenceRecord) bool {
	if !query.hasSnapshotOrControlSelectors() {
		return false
	}
	return !snapshotsMatch(query.Before, record.Before) || !snapshotsMatch(query.After, record.After) ||
		(!controlsEmpty(query.Controls) && !controlsEqual(query.Controls, record.Controls))
}
func (query InferenceQuery) hasSnapshotOrControlSelectors() bool {
	return query.Before.Source != "" || query.Before.Semantic != "" || query.After.Source != "" ||
		query.After.Semantic != "" || !controlsEmpty(query.Controls)
}
