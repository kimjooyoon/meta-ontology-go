package policycompilation

// EvaluateSourcePolicy applies the reduction rows emitted by the Gooo source.
// The evaluator supplies only typed evidence predicates; decisions, stages,
// steps, reasons, and UNKNOWN metadata come from the compiled source rows.
func EvaluateSourcePolicy(policy CompiledPolicy, input Case) DecisionResult {
	return evaluateRows(policy, input, sourceConditionMatches)
}

// IndependentEvaluate is a second interpreter used by the producer. It has a
// deliberately separate predicate implementation so equivalence is observed,
// not obtained by calling the source evaluator again.
func IndependentEvaluate(policy CompiledPolicy, input Case) DecisionResult {
	return evaluateRows(policy, input, independentConditionMatches)
}

type conditionMatcher func(string, string, string, Case) bool

func evaluateRows(policy CompiledPolicy, input Case, matches conditionMatcher) DecisionResult {
	result := baseResult(policy, input)
	if !safePolicyShape(policy) {
		return failClosed(result, "FIXED_DENOMINATOR_CHANGED")
	}
	for _, row := range policy.Reduction.Rules {
		if matches(row.Condition, policy.SourceDigest, policy.SemanticDigest, input) {
			return applyDecisionRule(result, row)
		}
	}
	return failClosed(result, "NO_REDUCTION_RULE_MATCHED")
}

func baseResult(policy CompiledPolicy, input Case) DecisionResult {
	return DecisionResult{
		CaseID: input.ID, PolicyDigest: policy.SourceDigest,
		SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator,
		BlockedBy: []string{},
	}
}

func safePolicyShape(policy CompiledPolicy) bool {
	return policy.Denominator == FixedDenominator && len(policy.Rules) == FixedDenominator &&
		policy.Reduction.Schema == ReductionSchema && len(policy.Reduction.Rules) == ReductionRuleCount
}

func sourceConditionMatches(condition, sourceDigest, semanticDigest string, input Case) bool {
	validSource := ValidDigest(input.ObservedSourceDigest)
	validArtifact := ValidDigest(input.ObservedArtifactSourceDigest)
	validIndependent := ValidDigest(input.ObservedIndependentDigest)
	validJudge := ValidDigest(input.ObservedGeneratedJudgeDigest)
	allValid := validSource && validArtifact && validIndependent && validJudge
	anyEmpty := input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedGeneratedJudgeDigest == "" || input.ObservedIndependentDigest == ""
	available := input.ProducerAvailable && input.ConsumerAvailable

	// These contradiction predicates are intentionally evaluated before
	// availability/UNKNOWN predicates. A valid contradiction is REFUTED first;
	// missing unrelated evidence cannot launder it into UNKNOWN.
	switch condition {
	case ConditionUnrecognizedTopDecision:
		return input.UpperDecision != "" && !knownDecision(input.UpperDecision)
	case ConditionSourceMismatch:
		return validSource && input.ObservedSourceDigest != sourceDigest
	case ConditionArtifactMismatch:
		return validSource && validArtifact && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest != sourceDigest
	case ConditionIndependentMismatch:
		return validSource && validArtifact && validIndependent && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest != semanticDigest
	case ConditionEvidenceUnavailable:
		return !available && !validSource && !validArtifact && !validIndependent && !validJudge
	case ConditionDigestUnavailable:
		return available && anyEmpty
	case ConditionMalformedDigest:
		return available && !anyEmpty && !allValid
	case ConditionSemanticEquivalence:
		return available && allValid && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest == semanticDigest
	default:
		return false
	}
}

// The independent matcher uses the same source-derived coordinates but keeps
// its predicate decisions structurally separate from sourceConditionMatches.
func independentConditionMatches(condition, sourceDigest, semanticDigest string, input Case) bool {
	sourceOK := ValidDigest(input.ObservedSourceDigest)
	artifactOK := ValidDigest(input.ObservedArtifactSourceDigest)
	independentOK := ValidDigest(input.ObservedIndependentDigest)
	judgeOK := ValidDigest(input.ObservedGeneratedJudgeDigest)
	complete := sourceOK && artifactOK && independentOK && judgeOK
	empty := input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedGeneratedJudgeDigest == "" || input.ObservedIndependentDigest == ""
	ready := input.ProducerAvailable && input.ConsumerAvailable

	switch condition {
	case ConditionUnrecognizedTopDecision:
		return input.UpperDecision != "" && input.UpperDecision != DecisionPass && input.UpperDecision != DecisionFailClosed && input.UpperDecision != DecisionUnknown
	case ConditionSourceMismatch:
		return sourceOK && input.ObservedSourceDigest != sourceDigest
	case ConditionArtifactMismatch:
		return sourceOK && artifactOK && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest != sourceDigest
	case ConditionIndependentMismatch:
		return sourceOK && artifactOK && independentOK && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest != semanticDigest
	case ConditionEvidenceUnavailable:
		return !ready && !sourceOK && !artifactOK && !independentOK && !judgeOK
	case ConditionDigestUnavailable:
		return ready && empty
	case ConditionMalformedDigest:
		return ready && !empty && !complete
	case ConditionSemanticEquivalence:
		return ready && complete && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest == semanticDigest
	default:
		return false
	}
}

func applyDecisionRule(result DecisionResult, rule DecisionRule) DecisionResult {
	result.Decision, result.Stage, result.Step, result.Reason = rule.Decision, rule.Stage, rule.Step, rule.Reason
	result.UnknownClass, result.NextOperation = rule.UnknownClass, rule.NextOperation
	result.BlockedBy = append([]string(nil), rule.BlockedBy...)
	if result.Decision != DecisionUnknown {
		result.UnknownClass, result.NextOperation = "", ""
		result.BlockedBy = []string{}
	}
	return result
}

func failClosed(result DecisionResult, reason string) DecisionResult {
	result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "COMPILE", 0, reason
	result.UnknownClass, result.NextOperation, result.BlockedBy = "", "", []string{}
	return result
}
