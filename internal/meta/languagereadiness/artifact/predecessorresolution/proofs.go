package predecessorresolution

func proofs(report Report, contiguous bool) []Proof {
	foundation := report.Selected != nil &&
		report.Summary.ValidCandidates == 1
	coherence := contiguous && report.Selected != nil &&
		report.Summary.ObservedAttempts == report.Summary.SelectedDepth+1
	regression := report.Selected != nil &&
		report.Summary.MissingAttempts == report.Summary.SelectedDepth &&
		report.Summary.AmbiguousCandidates == 0 &&
		report.Summary.RepositoryWrites == 0 &&
		report.Summary.ReadinessDeltaClaims == 0
	return []Proof{
		{ID: "exact-selected-baseline", Choice: "FOUNDATION", Passed: foundation},
		{ID: "contiguous-ancestry", Choice: "COHERENCE", Passed: coherence},
		{ID: "missing-only-read-only-skip", Choice: "REGRESSION", Passed: regression},
	}
}
