package predecessorbinding

import "fmt"

func Compare(before, after Report) (BindingTransition, error) {
	if err := Validate(before, before.HeadSHA); err != nil {
		return BindingTransition{}, fmt.Errorf("before binding invalid: %w", err)
	}
	if err := Validate(after, after.HeadSHA); err != nil {
		return BindingTransition{}, fmt.Errorf("after binding invalid: %w", err)
	}
	result := BindingTransition{Schema: BindingTransitionSchema,
		Decision: "LOWER_RESOLUTION", Reason: "BINDINGS_NOT_COMPARABLE",
		RegistryDigest: before.RegistryDigest, BeforeHeadSHA: before.HeadSHA,
		AfterHeadSHA: after.HeadSHA, Total: before.Summary.Total,
		BeforeStatic: before.Summary.StaticLiteral, AfterStatic: after.Summary.StaticLiteral,
		BeforeDynamic: before.Summary.DynamicInput, AfterDynamic: after.Summary.DynamicInput,
		BeforeBPS: before.Summary.DynamicBPS, AfterBPS: after.Summary.DynamicBPS,
		Unknown: before.Summary.Unknown + after.Summary.Unknown,
		RepositoryWrites: before.RepositoryWrites + after.RepositoryWrites}
	if before.RegistryDigest == after.RegistryDigest && before.Summary.Total == after.Summary.Total {
		result.Comparable = true
		result.StaticDelta = result.AfterStatic - result.BeforeStatic
		result.DynamicDelta = result.AfterDynamic - result.BeforeDynamic
		result.BPSDelta = result.AfterBPS - result.BeforeBPS
		classifyTransition(&result)
	}
	result.Indicators = transitionIndicators(result)
	result.Proofs = []Proof{
		{ID: "fixed-coordinate-registry", Choice: "FOUNDATION", Passed: result.Comparable},
		{ID: "integer-binding-arithmetic", Choice: "COHERENCE", Passed: result.Comparable},
		{ID: "resolved-binding-evidence", Choice: "REGRESSION", Passed: result.Unknown == 0},
		{ID: "read-only-transition", Choice: "FOUNDATION", Passed: result.RepositoryWrites == 0},
	}
	result.Digest = digestJSON(result)
	return result, nil
}

func classifyTransition(result *BindingTransition) {
	switch {
	case result.Unknown != 0 || result.RepositoryWrites != 0:
		result.Reason = "BINDING_GUARDRAIL_FAILED"
	case result.StaticDelta < 0 && result.DynamicDelta > 0 && result.BPSDelta > 0:
		result.Decision, result.Reason = "IMPROVED", "DYNAMIC_BINDING_PROVEN"
	case result.StaticDelta == 0 && result.DynamicDelta == 0 && result.BPSDelta == 0:
		result.Decision, result.Reason = "NO_CHANGE", "NO_NUMERIC_CHANGE"
	case result.StaticDelta > 0 || result.DynamicDelta < 0 || result.BPSDelta < 0:
		result.Decision, result.Reason = "REGRESSED", "STATIC_BINDING_REGRESSION"
	default:
		result.Reason = "POSITIVE_BINDING_DELTA_NOT_PROVEN"
	}
}
