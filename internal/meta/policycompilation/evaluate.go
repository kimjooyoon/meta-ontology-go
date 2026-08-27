package policycompilation

// IndependentEvaluate re-derives the result from the source-compiled
// reduction, not from the generated judge output. Its predicate dispatch is
// intentionally separate from sourceConditionMatches.
func IndependentEvaluate(policy CompiledPolicy, input Case) DecisionResult {
	result := baseResult(policy, input)
	if !safePolicyShape(policy) {
		return safetyFailure(result, "FIXED_DENOMINATOR_CHANGED")
	}
	for _, rule := range policy.Reduction.Rules {
		if independentConditionMatches(rule.Condition, policy.SourceDigest, input) {
			return applyDecisionRule(result, rule)
		}
	}
	return safetyFailure(result, "NO_REDUCTION_RULE_MATCHED")
}

func independentConditionMatches(condition, sourceDigest string, input Case) bool {
	available := input.ProducerAvailable && input.ConsumerAvailable
	switch condition {
	case ConditionEvidenceUnavailable:
		return !available
	case ConditionDigestUnavailable:
		return available && (input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedIndependentDigest == "")
	case ConditionSourceMismatch:
		return available && input.ObservedSourceDigest != "" && input.ObservedArtifactSourceDigest != "" && input.ObservedIndependentDigest != "" && input.ObservedSourceDigest != sourceDigest
	case ConditionArtifactMismatch:
		return available && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest != "" && input.ObservedArtifactSourceDigest != sourceDigest
	case ConditionIndependentMismatch:
		return available && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest != "" && input.ObservedIndependentDigest != sourceDigest
	case ConditionSemanticEquivalence:
		return available && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest == sourceDigest
	default:
		return false
	}
}

func sameDecision(left, right DecisionResult) bool {
	return left.CaseID == right.CaseID && left.Decision == right.Decision &&
		left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason &&
		left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest &&
		left.Denominator == right.Denominator
}
