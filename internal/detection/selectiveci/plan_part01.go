package selectiveci

func Plan(input Input) PlanResult {
	result := PlanResult{SchemaVersion: SchemaVersion, BaseSnapshotDigest: input.Base.Digest, HeadSnapshotDigest: input.Head.Digest}
	if err := input.Validate(); err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	changed, err := changedSemanticIDs(input.Base, input.Head)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	result.ChangedSemanticIDs = changed
	graph, err := buildGraph(input)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	coverage := EvaluateObligationCoverage(ObligationCoverageInput{
		SchemaVersion:  ObligationCoverageSchemaVersion,
		Graph:          graph,
		Registry:       input.Registry,
		SnapshotDigest: input.Head.Digest,
		ChangedRootIDs: changed,
	})
	if coverage.Decision != CoverageDecisionExact {
		return sealResult(fallback(result, string(coverage.Reason)))
	}
	required := coverage.RequiredObligationIDs
	commands, guards, err := selectedCommands(input.Registry, required)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	frontier, selected, err := commandFrontier(input, commands, guards)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	receiptDigests, pathIDs, err := validateSelectedEvidence(input, selected)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	result = fillSelection(result, selected, frontier)
	result.ResourceReceiptDigests = receiptDigests
	result.ProvenancePathIDs = pathIDs
	result.Status = StatusSelective
	result.ReasonCode = ReasonNone
	return sealResult(result)
}
func Evaluate(input Input) PlanResult { return Plan(input) }
func PlanJSON(data []byte) PlanResult {
	input, err := DecodeJSON(data)
	if err != nil {
		return sealResult(fallback(PlanResult{SchemaVersion: SchemaVersion}, reasonFor(err)))
	}
	return Plan(input)
}
func PlanJSONWithError(data []byte) (PlanResult, error) {
	input, err := DecodeJSON(data)
	if err != nil {
		return sealResult(fallback(PlanResult{SchemaVersion: SchemaVersion}, reasonFor(err))), err
	}
	return Plan(input), nil
}
func fallback(result PlanResult, reason string) PlanResult {
	result.Status = StatusFullSuiteFallback
	result.ReasonCode = reason
	result.SelectedCommandIDs = nil
	result.SelectedGuardCommandIDs = nil
	result.SelectedWorkIDs = nil
	result.ResourceReceiptDigests = nil
	result.ProvenancePathIDs = nil
	return result
}
