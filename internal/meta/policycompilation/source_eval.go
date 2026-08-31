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
		if sourceConditionMatches(rule.Condition, policy.SourceDigest, policy.SemanticDigest, input) {
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

func sourceConditionMatches(condition, sourceDigest, semanticDigest string, input Case) bool {
	available := input.ProducerAvailable && input.ConsumerAvailable
	switch condition {
	case ConditionEvidenceUnavailable:
		return !available
	case ConditionDigestUnavailable:
		return available && hasEmptyDigest(input)
	case ConditionMalformedDigest:
		return available && !hasEmptyDigest(input) && hasMalformedDigest(input)
	case ConditionSourceMismatch:
		return available && digestsValid(input) && input.ObservedSourceDigest != sourceDigest
	case ConditionArtifactMismatch:
		return available && digestsValid(input) && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest != sourceDigest
	case ConditionIndependentMismatch:
		return available && digestsValid(input) && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest != semanticDigest
	case ConditionSemanticEquivalence:
		return available && digestsValid(input) && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest == semanticDigest
	default:
		return false
	}
}

func hasEmptyDigest(input Case) bool {
	return input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedGeneratedJudgeDigest == "" || input.ObservedIndependentDigest == ""
}

func hasMalformedDigest(input Case) bool {
	return !ValidDigest(input.ObservedSourceDigest) || !ValidDigest(input.ObservedArtifactSourceDigest) || !ValidDigest(input.ObservedGeneratedJudgeDigest) || !ValidDigest(input.ObservedIndependentDigest)
}

func digestsValid(input Case) bool {
	return !hasEmptyDigest(input) && !hasMalformedDigest(input)
}

func applyDecisionRule(result DecisionResult, rule DecisionRule) DecisionResult {
	result.Decision, result.Stage, result.Step, result.Reason = rule.Decision, rule.Stage, rule.Step, rule.Reason
	return result
}

func safetyFailure(result DecisionResult, reason string) DecisionResult {
	result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "COMPILE", 3, reason
	return result
}
