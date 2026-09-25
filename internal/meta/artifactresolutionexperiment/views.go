package artifactresolutionexperiment

func buildViews(indicators []Indicator) []View {
	user := []string{"artifact.interface", "golden.interface", "replay.interface",
		"coherence.operation", "resolution.interface-definitions"}
	tool := []string{"artifact.manifest", "artifact.interface", "golden.manifest",
		"golden.interface", "replay.manifest", "replay.interface", "coherence.operation",
		"resolution.manifest-definitions", "resolution.interface-definitions", "compiler.emitter-registry"}
	governor := make([]string, len(indicators))
	for index, indicator := range indicators {
		governor[index] = indicator.ID
	}
	return []View{view("USER", "USER_VISIBLE", user, indicators),
		view("TOOL_AUTHOR", "TOOL_CONTRACT", tool, indicators),
		view("GOVERNOR", "FULL_RECEIPT", governor, indicators)}
}

func view(audience, resolution string, ids []string, indicators []Indicator) View {
	satisfied := 0
	for _, id := range ids {
		for _, indicator := range indicators {
			if indicator.ID == id && indicator.Satisfied {
				satisfied++
			}
		}
	}
	return View{Audience: audience, Resolution: resolution, Satisfied: satisfied,
		Total: len(ids), BasisPoints: basisPoints(satisfied, len(ids)), IndicatorIDs: ids}
}
