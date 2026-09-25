package query

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
func (budget *inferenceWorkBudget) edge() bool  { return budget.take(&budget.work.EdgesInspected) }
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
