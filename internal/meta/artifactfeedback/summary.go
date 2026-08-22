package artifactfeedback

func summarize(input Input) Summary {
	summary := Summary{RequiredInputs: 2, ExactHeadInputs: 1}
	if input.Coverage.CommitSHA == input.Cycle.HeadSHA {
		summary.ExactHeadInputs++
	} else {
		summary.StaleInputs++
	}
	if input.Coverage.Decision == "FIXED_POINT" || input.Coverage.Decision == "IMPROVE" {
		summary.BoundInputs++
	}
	if input.Cycle.Status == "BOUND" && input.Cycle.CIConclusion == "success" && !input.Cycle.PromotionAuthorized {
		summary.BoundInputs++
	}
	if input.CoverageReplayDigest == input.Coverage.ReportDigest {
		summary.ReplayBoundInputs++
	}
	if validBareDigest(input.Cycle.ReplayDigest) {
		summary.ReplayBoundInputs++
	}
	if (input.Coverage.Decision == "IMPROVE") != (input.Coverage.SelectedOperation != "") {
		summary.AmbiguousNextOperations++
	}
	summary.RepositoryWrites = input.RepositoryWrites + input.Coverage.Summary.RepositoryWrites
	minimum := min(summary.ExactHeadInputs, summary.BoundInputs, summary.ReplayBoundInputs)
	summary.ReadinessBasisPoints = minimum * 10000 / summary.RequiredInputs
	if summary.StaleInputs != 0 || summary.AmbiguousNextOperations != 0 || summary.RepositoryWrites != 0 {
		summary.ReadinessBasisPoints = 0
	}
	return summary
}

func min(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}
