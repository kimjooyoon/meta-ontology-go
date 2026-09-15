package toolchainconformance

func evaluateTamper(definitions []SurfaceDefinition, input Input,
	tamper TamperDefinition) CaseResult {
	result := CaseResult{ID: tamper.ID, Mutation: tamper.Mutation,
		Target: tamper.Target, ExpectedDecision: DecisionFailClosed,
		ObservedDecision: ResolutionLower, Status: "NOT_SATISFIED"}
	artifacts, err := applyTamper(input.Artifacts, tamper)
	if err != nil {
		result.EvidenceDigest = digestValue(err.Error())
		return result
	}
	summary, _ := inspectAll(definitions, artifacts, input.ExpectedHeadSHA)
	if blockingCount(summary) > 0 {
		result.ObservedDecision = DecisionFailClosed
		result.Status = "SATISFIED"
	} else {
		result.ObservedDecision = DecisionPass
	}
	result.EvidenceDigest = digestArtifacts(artifacts)
	return result
}
