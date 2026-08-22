package predecessorresolution

func indicators(report Report, contiguous bool) []Indicator {
	selected := report.Selected != nil
	depthBound := selected && report.Summary.SelectedDepth < SearchLimit
	attemptCardinality := selected && report.Summary.ObservedAttempts ==
		report.Summary.SelectedDepth+1
	missingCardinality := selected && report.Summary.MissingAttempts ==
		report.Summary.SelectedDepth
	return []Indicator{
		indicator("resolved-ancestor-cardinality", "OUTCOME",
			report.Summary.SelectedAncestors, "EQ", 1, selected),
		indicator("readiness-baseline-cardinality", "OUTCOME",
			boolInt(report.Selected != nil), "EQ", 1, selected),
		indicator("selected-depth-upper-bound", "DRIVER",
			report.Summary.SelectedDepth, "LT", SearchLimit, depthBound),
		indicator("ordered-attempt-cardinality", "DRIVER",
			report.Summary.ObservedAttempts, "EQ",
			report.Summary.SelectedDepth+1, attemptCardinality),
		indicator("contiguous-parent-links", "DRIVER",
			boolInt(contiguous), "EQ", 1, contiguous),
		indicator("missing-only-skip-cardinality", "DRIVER",
			report.Summary.MissingAttempts, "EQ",
			report.Summary.SelectedDepth, missingCardinality),
		indicator("valid-selected-candidate-cardinality", "GUARDRAIL",
			report.Summary.ValidCandidates, "EQ", 1,
			report.Summary.ValidCandidates == 1),
		indicator("ambiguous-candidate-cardinality", "GUARDRAIL",
			report.Summary.AmbiguousCandidates, "EQ", 0,
			report.Summary.AmbiguousCandidates == 0),
		indicator("repository-write-cardinality", "GUARDRAIL",
			report.Summary.RepositoryWrites, "EQ", 0,
			report.Summary.RepositoryWrites == 0),
		indicator("readiness-delta-claim-cardinality", "GUARDRAIL",
			report.Summary.ReadinessDeltaClaims, "EQ", 0,
			report.Summary.ReadinessDeltaClaims == 0),
	}
}

func indicator(id, kind string, value int, comparator string, target int,
	passed bool) Indicator {
	return Indicator{ID: id, Kind: kind, Value: value,
		Comparator: comparator, Target: target, Passed: passed}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
