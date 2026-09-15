package query

// Execute evaluates a request against a detached graph. It returns a sealed
// machine response and a Go error for rejected requests; neither path mutates
// the graph or promotes a candidate fact.
func (graph Graph) Execute(request Request) (Response, error) {
	normalized, err := request.Normalize()
	if err != nil {
		return graph.rejectedResponse(request, err)
	}
	requestHash, err := normalized.CanonicalDigest()
	if err != nil {
		return graph.rejectedResponse(normalized, err)
	}
	response := Response{
		Schema:      QueryEnvelopeSchema,
		Status:      ResponseOK,
		Request:     normalized,
		RequestHash: requestHash,
		Metadata:    envelopeMetadata(graph.Metadata()),
	}
	if normalized.Operation == OperationExact {
		response, err = graph.executeExact(response)
	} else if normalized.Operation == OperationDerived {
		response, err = graph.executeDerived(response)
	} else {
		response, err = graph.executeTraversal(response)
	}
	if err != nil {
		return graph.rejectedResponse(normalized, err)
	}
	if err := response.seal(); err != nil {
		return Response{}, err
	}
	return response, nil
}
func (graph Graph) executeDerived(response Response) (Response, error) {
	_, selection, err := normalizeLayer(response.Request.Layer)
	if err != nil {
		return response, err
	}
	result, err := graph.Derive(response.Request.Root, DerivedOptions{
		Rule: response.Request.Rule, MaxDepth: response.Request.MaxDepth,
		Limit: response.Request.Limit, Selection: selection,
	})
	if err != nil {
		return response, err
	}
	response.Metadata = envelopeMetadata(result.Metadata)
	response.Result.DerivedDeterministic, response.Result.DerivedCandidates = limitDerived(
		result.Deterministic, result.Candidates, response.Request.Limit,
	)
	return response, nil
}
func (graph Graph) executeExact(response Response) (Response, error) {
	_, selection, err := normalizeLayer(response.Request.Layer)
	if err != nil {
		return response, err
	}
	result, err := graph.ExactMatchWithOptions(
		NewExactQuery(response.Request.Root, response.Request.Relation, response.Request.Target),
		MatchOptions{Selection: selection},
	)
	if err != nil {
		return response, err
	}
	response.Metadata = envelopeMetadata(result.Metadata)
	response.Result.DeterministicMatches, response.Result.CandidateMatches = limitFacts(
		result.Deterministic, result.Candidates, response.Request.Limit,
	)
	return response, nil
}
