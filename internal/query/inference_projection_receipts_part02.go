package query

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func rejectedInferenceResponseWithWork(
	request InferenceQuery, requestHash string, work InferenceWork, err error,
) (InferenceQueryResult, error) {
	result := InferenceQueryResult{
		Schema: InferenceQuerySchema, Status: ResponseError, Request: request,
		RequestHash: requestHash, Work: work, Complete: false,
		Error: &EnvelopeError{Code: inferenceErrorCode(err), Message: err.Error()},
	}
	if sealErr := result.seal(); sealErr != nil {
		return InferenceQueryResult{}, sealErr
	}
	return result, err
}
func inferenceErrorCode(err error) string {
	if errors.Is(err, ErrInferenceQueryBudget) {
		return "inference_budget"
	}
	if errors.Is(err, ErrInferenceStaleSnapshot) {
		return "stale_snapshot"
	}
	if errors.Is(err, ErrInferenceUnsupportedPred) {
		return "unsupported_predicate"
	}
	if errors.Is(err, ErrInferenceChain) {
		return "invalid_chain"
	}
	if errors.Is(err, semantic.ErrInferencePath) {
		return "invalid_path"
	}
	if errors.Is(err, ErrInvalidInferenceQuery) {
		return "invalid_inference_query"
	}
	return "inference_query_rejected"
}
func sortInferenceRows(rows []InferenceRow) {
	sort.Slice(rows, func(i, j int) bool { return inferenceRowCanonical(rows[i]) < inferenceRowCanonical(rows[j]) })
}
func sortSemanticChangeRows(rows []SemanticChangeRow) {
	sort.Slice(rows, func(i, j int) bool { return semanticChangeRowCanonical(rows[i]) < semanticChangeRowCanonical(rows[j]) })
}
