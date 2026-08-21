package query

func (graph Graph) executeTraversal(response Response) (Response, error) {
	_, selection, err := normalizeLayer(response.Request.Layer)
	if err != nil {
		return response, err
	}
	result, err := graph.Traverse(response.Request.Root, TraversalOptions{
		Predicate: response.Request.Relation,
		Direction: envelopeDirection(response.Request.Direction),
		MaxDepth:  response.Request.MaxDepth,
		Limit:     response.Request.Limit,
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
