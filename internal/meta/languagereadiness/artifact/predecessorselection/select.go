package predecessorselection

func Select(input Input) (Result, error) {
	if len(input.Pagination.Pages) == 0 && input.Pagination.PageCount == 0 && input.Pagination.FailureReason == "" {
		// Hand-authored selection inputs from older receipts have no API page
		// surface. Treat that absence as a complete local fixture; live collection
		// always supplies at least one observed page or an explicit failure reason.
		input.Pagination.Complete = true
	}
	if err := validateInput(input); err != nil {
		return Result{}, err
	}
	report := Report{Schema: Schema, Repository: input.Repository,
		CurrentHeadSHA: input.CurrentHeadSHA, PredecessorSHA: input.PredecessorSHA,
		Decision: DecisionFailClosed, Reason: ReasonNotFound, Pagination: input.Pagination}
	report.Summary.ObservedCandidates = len(input.Candidates)
	var eligible []Result
	for _, candidate := range input.Candidates {
		if candidate.HeadSHA != input.PredecessorSHA {
			continue
		}
		report.Summary.ExactHeadCandidates++
		if !canonical(input, candidate) {
			continue
		}
		report.Summary.CanonicalCandidates++
		report.Summary.RepositoryWrites += candidate.RepositoryWrites
		if candidate.Conclusion == "success" {
			report.Summary.SuccessfulCandidates++
		}
		if !producerConformant(candidate) {
			continue
		}
		report.Summary.ProducerConformantCandidates++
		if candidate.ReadinessExpired || candidate.BindingExpired {
			continue
		}
		report.Summary.AvailableCandidates++
		selection, readinessRaw, bindingRaw, err := bindCandidate(candidate)
		if err != nil {
			continue
		}
		report.Summary.ValidCandidates++
		eligible = append(eligible, Result{Report: Report{Selected: &selection},
			BaselineRaw: readinessRaw, BindingRaw: bindingRaw})
	}
	if len(eligible) > 1 {
		report.Summary.AmbiguousCandidates = len(eligible) - 1
	}
	if !input.Pagination.Complete {
		report.Reason = input.Pagination.FailureReason
		if report.Reason == "" {
			report.Reason = ReasonPaginationIncomplete
		}
	} else {
		report.Reason = failureReason(report.Summary)
	}
	result := Result{Report: report}
	if report.Reason == "" {
		result = eligible[0]
		result.Report = report
		result.Report.Decision, result.Report.Reason = DecisionSelected, ReasonSelected
		result.Report.Selected = eligible[0].Report.Selected
	}
	result.Report.Proofs = proofs(result.Report)
	result.Report.ReportDigest = digestJSON(result.Report)
	return result, nil
}
