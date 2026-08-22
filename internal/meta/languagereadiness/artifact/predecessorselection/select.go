package predecessorselection

import "fmt"

func Select(input Input) (Result, error) {
	if err := validateInput(input); err != nil {
		return Result{}, err
	}
	report := Report{Schema: Schema, Repository: input.Repository,
		CurrentHeadSHA: input.CurrentHeadSHA, PredecessorSHA: input.PredecessorSHA,
		Decision: DecisionFailClosed, Reason: ReasonNotFound}
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
		if candidate.Conclusion != "success" {
			continue
		}
		report.Summary.SuccessfulCandidates++
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
	report.Reason = failureReason(report.Summary)
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

func validateInput(input Input) error {
	if input.Repository == "" || input.Branch == "" || input.Workflow == "" ||
		!validSHA(input.CurrentHeadSHA) || !validSHA(input.PredecessorSHA) ||
		input.CurrentHeadSHA == input.PredecessorSHA {
		return fmt.Errorf("readiness predecessor input identity malformed")
	}
	for _, candidate := range input.Candidates {
		if candidate.RunID <= 0 || candidate.RunAttempt <= 0 ||
			candidate.ReadinessArtifactID <= 0 || candidate.BindingArtifactID <= 0 ||
			candidate.RepositoryWrites < 0 || !validSHA(candidate.HeadSHA) {
			return fmt.Errorf("readiness predecessor candidate identity malformed")
		}
	}
	return nil
}
