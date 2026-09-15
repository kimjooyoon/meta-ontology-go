package selfimprovementobservation

func validSourceIndicators(indicators []SourceIndicator) bool {
	seen := map[string]bool{}
	for _, indicator := range indicators {
		if seen[indicator.ID] || indicator.ID == "" || indicator.MetaOperation == "" ||
			!validProofChoice(indicator.ProofChoice) || !indicator.Satisfied {
			return false
		}
		seen[indicator.ID] = true
	}
	return len(indicators) == 15
}

func validSourceProofs(report SourceReport) bool {
	seen := map[string]bool{}
	for _, proof := range report.Proofs {
		if seen[proof.Choice] || !validProofChoice(proof.Choice) || proof.MetaOperation == "" ||
			!proof.Passed || proof.EvidenceDigest != report.FactsDigest {
			return false
		}
		seen[proof.Choice] = true
	}
	return len(report.Proofs) == 3 && len(seen) == 3
}

func validSourceViews(views []SourceView) bool {
	expected := map[string]int{"USER": 6, "TOOL_AUTHOR": 12, "GOVERNOR": 15}
	for _, view := range views {
		total, ok := expected[view.Audience]
		if !ok || view.Total != total || view.Satisfied != total || view.BasisPoints != 10000 ||
			len(view.IndicatorIDs) != total || view.Resolution == "" {
			return false
		}
		delete(expected, view.Audience)
	}
	return len(views) == 3 && len(expected) == 0
}

func validProofChoice(choice string) bool {
	return choice == "FOUNDATION" || choice == "COHERENCE" || choice == "REGRESSION"
}
