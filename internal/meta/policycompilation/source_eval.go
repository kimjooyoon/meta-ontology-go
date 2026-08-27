package policycompilation

// EvaluateSourcePolicy interprets the source-bound semantic contract directly.
// It is a third observation: the generated judge is executable code, while
// this path reads the compiled policy contract and applies its fail-closed
// decision table without consulting either execution result.
func EvaluateSourcePolicy(policy CompiledPolicy, input Case) DecisionResult {
	result := DecisionResult{
		CaseID: input.ID, PolicyDigest: policy.SourceDigest,
		SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator,
	}
	if policy.Denominator != FixedDenominator || len(policy.Rules) != FixedDenominator {
		result.Decision, result.Stage, result.Step, result.Reason = DecisionFailClosed, "COMPILE", 3, "FIXED_DENOMINATOR_CHANGED"
		return result
	}
	if input.ProducerAvailable && input.ConsumerAvailable {
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
	result.Decision, result.Stage, result.Step, result.Reason = DecisionUnknown, "VERIFY", 4, "EVIDENCE_UNAVAILABLE"
	return result
}
