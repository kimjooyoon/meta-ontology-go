package query

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
