package userjourneyscorecard

func buildViews(indicators []Indicator) []AudienceView {
	specs := []struct {
		audience, resolution string
		ids                  []string
	}{
		{"USER", "USER_VISIBLE", []string{"functional.user-journeys", "resource.envelopes", "profile.output-replay", "guardrail.wall", "guardrail.rss", "guardrail.binary-size"}},
		{"TOOL_AUTHOR", "TOOL_CONTRACT", []string{"functional.cli-contract", "functional.user-journeys", "resource.envelopes", "profile.output-replay", "functional.structured-output", "functional.language-operations", "binding.meta-operations", "guardrail.wall", "guardrail.rss", "guardrail.binary-size"}},
		{"GOVERNOR", "FULL_RECEIPT", indicatorIDs(indicators)},
	}
	views := make([]AudienceView, 0, len(specs))
	for _, spec := range specs {
		satisfied := 0
		for _, id := range spec.ids {
			for _, indicator := range indicators {
				if indicator.ID == id && indicator.Satisfied {
					satisfied++
				}
			}
		}
		decision := "FAIL_CLOSED"
		if satisfied == len(spec.ids) {
			decision = "PASS"
		}
		views = append(views, AudienceView{Audience: spec.audience, Decision: decision,
			Resolution: spec.resolution, Satisfied: satisfied, Total: len(spec.ids),
			BasisPoints: satisfied * 10000 / len(spec.ids), IndicatorIDs: spec.ids})
	}
	return views
}

func indicatorIDs(indicators []Indicator) []string {
	ids := make([]string, len(indicators))
	for index, indicator := range indicators {
		ids[index] = indicator.ID
	}
	return ids
}

func proofs(indicators []Indicator) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", Claim: "exact head, corpus, executable, source, and meta operations are bound", Evidence: digestJSON(indicators), Passed: proofPassed(indicators, "FOUNDATION")},
		{Choice: "COHERENCE", Claim: "user outputs and resource reductions agree with the functional receipt", Evidence: digestJSON(indicators), Passed: proofPassed(indicators, "COHERENCE")},
		{Choice: "REGRESSION", Claim: "resource ceilings, replay, and repository effects remain bounded", Evidence: digestJSON(indicators), Passed: proofPassed(indicators, "REGRESSION")},
	}
}

func proofPassed(indicators []Indicator, choice string) bool {
	for _, indicator := range indicators {
		if indicator.ProofChoice == choice && !indicator.Satisfied {
			return false
		}
	}
	return true
}
