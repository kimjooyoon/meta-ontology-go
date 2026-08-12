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

func (graph Graph) executeTraversal(response Response) (Response, error) {
	_, selection, err := normalizeLayer(response.Request.Layer)
	if err != nil {
		return response, err
	}
	result, err := graph.Traverse(response.Request.Root, TraversalOptions{
		Predicate: response.Request.Relation,
		Direction: envelopeDirection(response.Request.Direction),
		MaxDepth:  response.Request.MaxDepth,
		Selection: selection,
	})
	if err != nil {
		return response, err
	}
	response.Metadata = envelopeMetadata(result.Metadata)
	response.Result.DeterministicPaths, response.Result.CandidatePaths = limitPaths(
		result.Deterministic, result.Candidates, response.Request.Limit,
	)
	return response, nil
}

func envelopeDirection(direction string) Direction {
	switch direction {
	case "incoming":
		return Incoming
	case "both":
		return Both
	default:
		return Outgoing
	}
}

func limitFacts(deterministic, candidates []Fact, limit int) ([]Fact, []Fact) {
	if len(deterministic) > limit {
		return append([]Fact(nil), deterministic[:limit]...), nil
	}
	remaining := limit - len(deterministic)
	if len(candidates) > remaining {
		candidates = candidates[:remaining]
	}
	return append([]Fact(nil), deterministic...), append([]Fact(nil), candidates...)
}

func limitPaths(deterministic, candidates []Path, limit int) ([]Path, []Path) {
	if len(deterministic) > limit {
		return append([]Path(nil), deterministic[:limit]...), nil
	}
	remaining := limit - len(deterministic)
	if len(candidates) > remaining {
		candidates = candidates[:remaining]
	}
	return append([]Path(nil), deterministic...), append([]Path(nil), candidates...)
}

func limitDerived(deterministic, candidates []DerivedFact, limit int) ([]DerivedFact, []DerivedFact) {
	if len(deterministic) > limit {
		return append([]DerivedFact(nil), deterministic[:limit]...), nil
	}
	remaining := limit - len(deterministic)
	if len(candidates) > remaining {
		candidates = candidates[:remaining]
	}
	return append([]DerivedFact(nil), deterministic...), append([]DerivedFact(nil), candidates...)
}

func (graph Graph) rejectedResponse(request Request, err error) (Response, error) {
	response := Response{
		Schema:   QueryEnvelopeSchema,
		Status:   ResponseError,
		Request:  request,
		Metadata: envelopeMetadata(graph.Metadata()),
		Error:    &EnvelopeError{Code: errorCode(err), Message: err.Error()},
	}
	if normalized, normalizeErr := request.Normalize(); normalizeErr == nil {
		response.Request = normalized
		response.RequestHash, _ = normalized.CanonicalDigest()
	}
	if sealErr := response.seal(); sealErr != nil {
		return Response{}, sealErr
	}
	return response, err
}
