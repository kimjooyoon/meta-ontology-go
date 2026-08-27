package policycompilation

// IndependentEvaluate is deliberately separate from the generated judge
// source. It re-derives the result from evidence, not from the judge output.
func IndependentEvaluate(policy CompiledPolicy, input Case) DecisionResult {
	result := DecisionResult{
		CaseID: input.ID, PolicyDigest: policy.SourceDigest,
		SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator,
	}
	if !input.ProducerAvailable || !input.ConsumerAvailable {
		result.Decision, result.Stage, result.Step, result.Reason = DecisionUnknown, "VERIFY", 4, "EVIDENCE_UNAVAILABLE"
		return result
	}
	if input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedIndependentDigest == "" {
		result.Decision, result.Stage, result.Step, result.Reason = DecisionUnknown, "VERIFY", 4, "DIGEST_UNAVAILABLE"
		return result
	}
	if input.ObservedSourceDigest != policy.SourceDigest {
		result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "REDUCE", 7, "SOURCE_DIGEST_MISMATCH"
		return result
	}
	if input.ObservedArtifactSourceDigest != policy.SourceDigest {
		result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "CONSUME", 2, "ARTIFACT_SOURCE_MISMATCH"
		return result
	}
	if input.ObservedIndependentDigest != policy.SourceDigest {
		result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "VERIFY", 4, "INDEPENDENT_SOURCE_MISMATCH"
		return result
	}
	result.Decision, result.Stage, result.Step, result.Reason = DecisionPass, "REDUCE", 7, "SEMANTIC_EQUIVALENCE_PROVED"
	return result
}

func sameDecision(left, right DecisionResult) bool {
	return left.CaseID == right.CaseID && left.Decision == right.Decision &&
		left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason &&
		left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest &&
		left.Denominator == right.Denominator
}
