package workfrontier

// EvaluateR4 is the bounded detector entry point. It never consults the
// legacy selector's implicit defaults.
func EvaluateR4(input R4Input) R4Result {
	input = normalizeR4Input(input)
	if !r4RequiredInputKnown(input) {
		return r4RequiredInputResult()
	}
	if reason := validateR4Bindings(input); reason != "" {
		if reason == R4ReasonMalformedBinding {
			return r4FailClosed(r4Graph{}, reason)
		}
		return r4Unknown(r4Graph{}, reason)
	}
	graph, graphReason := buildR4Graph(input)
	if graphReason != "" {
		return r4FailClosed(graph, graphReason)
	}
	if !r4StableDeclarationsKnown(input) {
		return r4Unknown(graph, R4ReasonRequiredInputMissing)
	}
	if reason := validateR4Rules(graph, input.Rules); reason != "" {
		if reason == R4ReasonDuplicateSCCRule || reason == R4ReasonConflictingSCCRule {
			return r4FailClosed(graph, reason)
		}
		return r4Unknown(graph, reason)
	}
	legacy := r4LegacyInput(input, graph)
	result := r4ResultFromGraph(graph)
	ready, result := classifyR4Paths(legacy, graph, result)
	result = selectR4Paths(input, legacy, ready, result)
	return finishR4Result(result)
}
func r4RequiredInputResult() R4Result {
	return R4Result{
		SchemaVersion:     R4SchemaVersion,
		Status:            R4StatusUnknown,
		Reason:            R4ReasonRequiredInputMissing,
		Quality:           R4StatusUnknown,
		FullSuiteRequired: true,
	}
}
func r4LegacyInput(input R4Input, graph r4Graph) Input {
	return Input{
		SchemaVersion:            SchemaVersion,
		SnapshotDigest:           input.SnapshotDigest,
		PolicyDigest:             input.PolicyDigest,
		RegistryDigest:           input.RegistryDigest,
		MinimumSelectedPressures: input.MinimumSelectedPressures,
		Capacity:                 input.Capacity,
		Pressures:                append([]Pressure(nil), input.Pressures...),
		States:                   append([]ObligationState(nil), input.States...),
		Paths:                    append([]RepairPath(nil), graph.reachablePaths...),
	}
}
