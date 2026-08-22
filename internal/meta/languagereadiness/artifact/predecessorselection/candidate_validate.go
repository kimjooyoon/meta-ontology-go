package predecessorselection

import "fmt"

func validateInput(input Input) error {
	if input.Repository == "" || input.Branch == "" || input.Workflow == "" ||
		!validSHA(input.CurrentHeadSHA) || !validSHA(input.PredecessorSHA) ||
		input.CurrentHeadSHA == input.PredecessorSHA {
		return fmt.Errorf("readiness predecessor input identity malformed")
	}
	for _, candidate := range input.Candidates {
		if candidate.RunID <= 0 || candidate.RunAttempt <= 0 ||
			candidate.ReadinessArtifactID <= 0 || candidate.BindingArtifactID <= 0 ||
			candidate.RepositoryWrites < 0 || candidate.ProducerJobMatches < 0 ||
			!validSHA(candidate.HeadSHA) {
			return fmt.Errorf("readiness predecessor candidate identity malformed")
		}
		if candidate.ProducerJobMatches == 1 &&
			(candidate.ProducerJobID <= 0 || candidate.ProducerJobRunAttempt <= 0 ||
				candidate.ProducerJobName == "") {
			return fmt.Errorf("readiness predecessor producer identity malformed")
		}
	}
	return nil
}

func producerConformant(candidate Candidate) bool {
	return candidate.ProducerJobMatches == 1 && candidate.ProducerJobID > 0 &&
		candidate.ProducerJobRunAttempt == candidate.RunAttempt &&
		candidate.ProducerJobName == ProducerJobName &&
		candidate.ProducerJobStatus == "completed" &&
		candidate.ProducerJobConclusion == "success"
}
