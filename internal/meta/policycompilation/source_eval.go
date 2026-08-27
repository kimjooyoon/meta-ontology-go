package policycompilation

// EvaluateSourcePolicy interprets the reduction rows emitted by the Gooo
// source. The source supplies every output value; this function only provides
// the typed evidence predicates that the schema permits.
func EvaluateSourcePolicy(policy CompiledPolicy, input Case) DecisionResult {
	result := baseResult(policy, input)
	if !safePolicyShape(policy) {
		return safetyFailure(result, "FIXED_DENOMINATOR_CHANGED")
	}
	for _, rule := range policy.Reduction.Rules {
		if sourceConditionMatches(rule.Condition, policy.SourceDigest, input) {
			return applyDecisionRule(result, rule)
		}
	}
	return safetyFailure(result, "NO_REDUCTION_RULE_MATCHED")
}

func baseResult(policy CompiledPolicy, input Case) DecisionResult {
	return DecisionResult{CaseID: input.ID, PolicyDigest: policy.SourceDigest, SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator}
}

func safePolicyShape(policy CompiledPolicy) bool {
	return policy.Denominator == FixedDenominator && len(policy.Rules) == FixedDenominator && policy.Reduction.Schema == ReductionSchema && len(policy.Reduction.Rules) == ReductionRuleCount
}

func sourceConditionMatches(condition, sourceDigest string, input Case) bool {
	switch condition {
	case ConditionEvidenceUnavailable:
		return !input.ProducerAvailable || !input.ConsumerAvailable
	case ConditionDigestUnavailable:
		return input.ProducerAvailable && input.ConsumerAvailable && (input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedIndependentDigest == "")
	case ConditionSourceMismatch:
		return input.ProducerAvailable && input.ConsumerAvailable && input.ObservedSourceDigest != "" && input.ObservedArtifactSourceDigest != "" && input.ObservedIndependentDigest != "" && input.ObservedSourceDigest != sourceDigest
	case ConditionArtifactMismatch:
		return input.ProducerAvailable && input.ConsumerAvailable && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest != "" && input.ObservedArtifactSourceDigest != sourceDigest
	case ConditionIndependentMismatch:
		return input.ProducerAvailable && input.ConsumerAvailable && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest != "" && input.ObservedIndependentDigest != sourceDigest
	case ConditionSemanticEquivalence:
		return input.ProducerAvailable && input.ConsumerAvailable && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest == sourceDigest
	default:
		return false
	}
}

func applyDecisionRule(result DecisionResult, rule DecisionRule) DecisionResult {
	result.Decision, result.Stage, result.Step, result.Reason = rule.Decision, rule.Stage, rule.Step, rule.Reason
	return result
}

func safetyFailure(result DecisionResult, reason string) DecisionResult {
	result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "COMPILE", 3, reason
	return result
}
