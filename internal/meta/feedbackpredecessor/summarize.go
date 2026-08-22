package feedbackpredecessor

func summarize(input Input) (Summary, []Candidate) {
	summary := Summary{ObservedCandidates: len(input.Candidates)}
	expectedName := "artifact-feedback-resolution-" + input.PredecessorSHA
	var eligible []Candidate
	for _, candidate := range input.Candidates {
		if candidate.HeadSHA != input.PredecessorSHA {
			continue
		}
		summary.ExactHeadCandidates++
		if candidate.HeadBranch != input.CanonicalBranch ||
			candidate.Workflow != input.CanonicalWorkflow ||
			candidate.Event != "push" || candidate.ArtifactName != expectedName {
			continue
		}
		summary.CanonicalCandidates++
		summary.RepositoryWrites += candidate.RepositoryWrites
		if candidate.Conclusion != "success" {
			continue
		}
		summary.SuccessfulCandidates++
		if candidate.Expired {
			continue
		}
		summary.AvailableCandidates++
		if !validDigest(candidate.ReceiptDigest) {
			continue
		}
		summary.ReceiptBoundCandidates++
		eligible = append(eligible, candidate)
	}
	if len(eligible) > 1 {
		summary.AmbiguousCandidates = len(eligible) - 1
	}
	return summary, eligible
}
