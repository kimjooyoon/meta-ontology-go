package predecessorselection

import "fmt"

func validateInput(input Input) error {
	if input.Repository == "" || input.Branch == "" || input.Workflow == "" ||
		!validSHA(input.CurrentHeadSHA) || !validSHA(input.PredecessorSHA) ||
		input.CurrentHeadSHA == input.PredecessorSHA {
		return fmt.Errorf("readiness predecessor input identity malformed")
	}
	if input.Pagination.PageCount != len(input.Pagination.Pages) || input.Pagination.PageCount < 0 {
		return fmt.Errorf("readiness predecessor pagination inventory malformed")
	}
	seenPages := map[string]struct{}{}
	for _, page := range input.Pagination.Pages {
		if page.EndpointClass == "" || page.URL == "" || page.PageNumber < 1 || (page.HTTPStatus != 0 && (page.HTTPStatus < 100 || page.HTTPStatus > 599)) || page.BodyDigest == "" || page.BodyBytes < 0 || page.NextLinkDigest == "" {
			return fmt.Errorf("readiness predecessor pagination page malformed")
		}
		if _, exists := seenPages[page.URL]; exists {
			return fmt.Errorf("readiness predecessor pagination URL duplicated")
		}
		seenPages[page.URL] = struct{}{}
	}
	if !input.Pagination.Complete && input.Pagination.FailureReason == "" {
		return fmt.Errorf("incomplete pagination requires a failure reason")
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
