package feedbackpredecessor

import "fmt"

func Select(input Input) (Report, error) {
	if err := validate(input); err != nil {
		return Report{}, err
	}
	summary, eligible := summarize(input)
	report := Report{Schema: Schema, Repository: input.Repository,
		PredecessorSHA: input.PredecessorSHA, Summary: summary}
	decide(&report, eligible)
	report.Indicators = indicators(report)
	report.ReportDigest = digestJSON(report)
	return report, nil
}

func validate(input Input) error {
	if input.Repository == "" || input.CanonicalBranch == "" ||
		input.CanonicalWorkflow == "" || !validSHA(input.PredecessorSHA) {
		return fmt.Errorf("predecessor selection identity is malformed")
	}
	for _, candidate := range input.Candidates {
		if candidate.ArtifactID <= 0 || candidate.RunID <= 0 || candidate.RunAttempt <= 0 ||
			candidate.ArtifactName == "" || candidate.HeadBranch == "" ||
			candidate.Workflow == "" || candidate.Event == "" ||
			candidate.Conclusion == "" || !validSHA(candidate.HeadSHA) ||
			candidate.RepositoryWrites < 0 {
			return fmt.Errorf("predecessor candidate identity is malformed")
		}
	}
	return nil
}
