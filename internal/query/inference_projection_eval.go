package query

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type inferenceWorkBudget struct {
	work InferenceWork
}

func newInferenceWorkBudget(limit int) *inferenceWorkBudget {
	return &inferenceWorkBudget{work: InferenceWork{Limit: limit}}
}

func (budget *inferenceWorkBudget) take(kind *int) bool {
	if budget.work.Used >= budget.work.Limit {
		return false
	}
	budget.work.Used++
	(*kind)++
	return true
}

func (budget *inferenceWorkBudget) edge() bool { return budget.take(&budget.work.EdgesInspected) }

func (budget *inferenceWorkBudget) claim() bool { return budget.take(&budget.work.ClaimsInspected) }

func (budget *inferenceWorkBudget) evidence() bool {
	return budget.take(&budget.work.EvidenceInspected)
}

func (budget *inferenceWorkBudget) chain() bool { return budget.take(&budget.work.ChainInspected) }

// Execute evaluates one explicit bounded request over the detached snapshot.
// It never writes the path, a Graph, or a semantic authority record.
func (projection InferenceProjection) Execute(request InferenceQuery) (InferenceQueryResult, error) {
	normalized, err := request.normalized()
	if err != nil {
		return rejectedInferenceResponse(request, err)
	}
	requestHash, err := normalized.CanonicalDigest()
	if err != nil {
		return rejectedInferenceResponse(normalized, err)
	}
	budget := newInferenceWorkBudget(normalized.MaxWork)
	edges, matchedEdges, selectedEvidence, err := projection.scanEdges(normalized, budget)
	if err != nil {
		return projection.scanFailure(normalized, requestHash, budget, err)
	}
	claims, err := projection.scanClaims(normalized, budget, edges, selectedEvidence)
	if err != nil {
		return projection.scanFailure(normalized, requestHash, budget, err)
	}
	evidence, err := projection.scanEvidence(normalized, budget, edges, claims, selectedEvidence)
	if err != nil {
		return projection.scanFailure(normalized, requestHash, budget, err)
	}
	return projection.finish(normalized, requestHash, budget, edges, claims, evidence, matchedEdges)
}

func (projection InferenceProjection) scanFailure(
	request InferenceQuery, requestHash string, budget *inferenceWorkBudget, err error,
) (InferenceQueryResult, error) {
	return rejectedInferenceResponseWithWork(request, requestHash, budget.work, err)
}

func (projection InferenceProjection) finish(
	request InferenceQuery, requestHash string, budget *inferenceWorkBudget,
	edges []InferenceRow, claims []SemanticChangeRow, evidence []InferenceEvidenceRow,
	matchedEdges []semantic.InferenceEdge,
) (InferenceQueryResult, error) {
	result := InferenceQueryResult{
		Schema: InferenceQuerySchema, Status: ResponseOK, Request: request,
		RequestHash: requestHash, Edges: edges, Claims: claims, Evidence: evidence,
		Work: budget.work, Complete: false,
	}
	if request.Explain {
		chain, err := projection.explanationChain(request, matchedEdges, edges, budget)
		if err != nil {
			return projection.chainFailure(request, requestHash, budget, err)
		}
		result.Chain = &chain
	}
	result.Work = budget.work
	result.Complete = true
	if err := result.seal(); err != nil {
		return InferenceQueryResult{}, err
	}
	return result, nil
}

// Query and Evaluate are concise aliases for Execute.
func (projection InferenceProjection) Query(request InferenceQuery) (InferenceQueryResult, error) {
	return projection.Execute(request)
}

func (projection InferenceProjection) Evaluate(request InferenceQuery) (InferenceQueryResult, error) {
	return projection.Execute(request)
}

// QueryInferencePath validates the path and returns an explicit non-success
// response for invalid path input rather than treating it as empty data.
func QueryInferencePath(path semantic.InferencePathV1, request InferenceQuery) (InferenceQueryResult, error) {
	projection, err := NewInferenceProjection(path)
	if err != nil {
		return rejectedInferenceResponse(request, fmt.Errorf("%w: %v", semantic.ErrInferencePath, err))
	}
	return projection.Execute(request)
}

func QueryInference(path semantic.InferencePathV1, request InferenceQuery) (InferenceQueryResult, error) {
	return QueryInferencePath(path, request)
}

func (graph Graph) QueryInferencePath(path semantic.InferencePathV1, request InferenceQuery) (InferenceQueryResult, error) {
	return QueryInferencePath(path, request)
}

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
