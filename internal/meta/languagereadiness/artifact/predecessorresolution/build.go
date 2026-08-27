package predecessorresolution

import "fmt"

func Build(input Input) (Report, error) {
	if err := validateInput(input); err != nil {
		return Report{}, err
	}
	report := Report{Schema: Schema, Repository: input.Repository,
		CurrentHeadSHA:          input.CurrentHeadSHA,
		ImmediatePredecessorSHA: input.ImmediatePredecessorSHA,
		Decision:                DecisionFailClosed, Reason: ReasonExhausted,
		Conformance:             ConformancePass, Resolution: ResolutionLower}
	report.Summary.SelectedDepth = -1
	contiguous := true
	for index, attempt := range input.Attempts {
		kind, err := validateAttempt(input, index, attempt)
		if err != nil {
			return Report{}, err
		}
		report.Attempts = append(report.Attempts, AttemptReceipt{Depth: attempt.Depth,
			AncestorSHA: attempt.AncestorSHA, ParentSHA: attempt.ParentSHA,
			Selection: attempt.Selection.Report})
		addAttemptSummary(&report.Summary, attempt, kind)
		if index > 0 && input.Attempts[index-1].ParentSHA != attempt.AncestorSHA {
			contiguous = false
		}
		if kind == attemptSelected {
			resolve(&report, attempt)
		} else if kind == attemptBlocked {
			report.Reason = ReasonBlocked
			report.BlockingSelectionReason = attempt.Selection.Report.Reason
		}
	}
	if report.Selected == nil && report.Reason == ReasonExhausted &&
		len(input.Attempts) != SearchLimit {
		return Report{}, fmt.Errorf("ancestor resolution ended before fixed limit")
	}
	finalize(&report, contiguous)
	if err := Validate(report); err != nil {
		return Report{}, err
	}
	return report, nil
}

func resolve(report *Report, attempt Attempt) {
	selected := attempt.Selection.Report.Selected
	report.Decision, report.Reason = DecisionResolved, ReasonResolved
	report.Resolution = ResolutionExact
	report.Summary.SelectedAncestors = 1
	report.Summary.SelectedDepth = attempt.Depth
	report.Selected = &Resolution{Depth: attempt.Depth,
		AncestorSHA:     attempt.AncestorSHA,
		SelectionDigest: attempt.Selection.Report.ReportDigest,
		Baseline:        selected.Baseline}
}

func finalize(report *Report, contiguous bool) {
	if report.BlockingSelectionReason == "" {
		value := 0
		report.Summary.ReadinessDeltaClaims = &value
	}
	report.Summary.ObservedAttempts = len(report.Attempts)
	report.Summary.SearchLimit = SearchLimit
	report.Indicators = indicators(*report, contiguous)
	for _, indicator := range report.Indicators {
		if indicator.Passed {
			report.Summary.CoordinatesCompleted++
		}
	}
	report.Summary.CoordinatesTotal = len(report.Indicators)
	report.Summary.BasisPoints = report.Summary.CoordinatesCompleted * 10000 /
		report.Summary.CoordinatesTotal
	report.Proofs = proofs(*report, contiguous)
	report.ReportDigest = digestJSON(*report)
}
