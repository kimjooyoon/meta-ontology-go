package languageexampleexperiment

func views(values []Indicator) []View {
	user := []string{
		"value.primary-artifact", "value.golden-match", "value.deterministic-replay",
		"guardrail.wall", "guardrail.rss", "guardrail.binary",
	}
	tool := append(append([]string{}, user...),
		"value.artifact-digest-integrity", "compiler.source-files", "compiler.gooo-definition-bps",
		"compiler.emitter-registry", "resource.samples", "resource.valid-samples")
	governor := make([]string, len(values))
	for index, value := range values {
		governor[index] = value.ID
	}
	return []View{
		makeView("USER", "USER_VISIBLE", user, values),
		makeView("TOOL_AUTHOR", "TOOL_CONTRACT", tool, values),
		makeView("GOVERNOR", "FULL_RECEIPT", governor, values),
	}
}

func makeView(audience, resolution string, ids []string, values []Indicator) View {
	satisfied := 0
	for _, id := range ids {
		for _, value := range values {
			if value.ID == id && value.Satisfied {
				satisfied++
			}
		}
	}
	basisPoints := 0
	if len(ids) > 0 {
		basisPoints = satisfied * 10000 / len(ids)
	}
	return View{Audience: audience, Resolution: resolution, Satisfied: satisfied,
		Total: len(ids), BasisPoints: basisPoints, IndicatorIDs: ids}
}
