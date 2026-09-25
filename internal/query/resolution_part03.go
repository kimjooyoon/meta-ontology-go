package query

import (
	"fmt"
)

// ResolveGeneratedCode evaluates the fixed two-hop semantic resolution view.
// It reads detached graph snapshots only and never inserts a derived row.
func (graph Graph) ResolveGeneratedCode(request ResolutionRequest) (ResolutionResponse, error) {
	normalized, selection, err := request.Normalize()
	if err != nil {
		return graph.rejectedResolution(request, err)
	}
	requestHash, err := normalized.CanonicalDigest()
	if err != nil {
		return graph.rejectedResolution(normalized, err)
	}
	root, ok := graph.Node(normalized.Business)
	if !ok {
		return graph.rejectedResolution(normalized, fmt.Errorf("%w: %s", ErrUnknownEndpoint, normalized.Business))
	}
	if root.Kind != EntityNodeKind {
		return graph.rejectedResolution(normalized, envelopeError(
			ErrInvalidResolution, "invalid_business_kind", "business endpoint must be an Entity",
		))
	}
	deterministic, candidates := graph.resolutionRows(
		normalized.Business, selection, normalized.Limit,
	)
	status := StatusDeferred
	if len(deterministic)+len(candidates) > 0 {
		status = ResolutionResolved
	}
	metadata := resolutionMetadata(graph.Metadata(), status)
	response := ResolutionResponse{
		Schema: ResolutionSchema, Status: status, Request: normalized,
		RequestHash: requestHash, Deterministic: deterministic,
		Candidates: candidates, Metadata: metadata,
	}
	if err := response.seal(); err != nil {
		return ResolutionResponse{}, err
	}
	return response, nil
}
func (graph Graph) resolutionRows(root ID, selection FactSelection, limit int) ([]ResolutionRow, []ResolutionRow) {
	facts := graph.AllFacts()
	deterministic := graph.resolutionLayer(root, facts, SelectDeterministic, limit, nil)
	if selection == SelectDeterministic || len(deterministic) >= limit {
		return deterministic, nil
	}
	remaining := limit - len(deterministic)
	if selection == SelectCandidate {
		return nil, graph.resolutionLayer(root, facts, SelectCandidate, remaining, nil)
	}
	blocked := resolutionKeys(deterministic)
	candidates := graph.resolutionLayer(root, facts, SelectAll, remaining, blocked)
	return deterministic, candidates
}
