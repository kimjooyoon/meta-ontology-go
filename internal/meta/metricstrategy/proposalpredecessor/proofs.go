package proposalpredecessor

func buildProofs(selected *Selected, predecessorSHA string, summary Summary) ([]Proof, error) {
	value := Selected{}
	if selected != nil {
		value = *selected
	}
	specs := []struct {
		id, choice, operation string
		passed                bool
		evidence              any
	}{
		{"exact-predecessor-head", "FOUNDATION", "bind-exact-proposal-predecessor", selected != nil && value.HeadSHA == predecessorSHA, []string{value.HeadSHA, predecessorSHA}},
		{"canonical-synthesis-job", "COHERENCE", "bind-canonical-proposal-synthesis", selected != nil && canonicalSelected(value), []any{value.RunID, value.RunAttempt, value.Event, value.Status, value.Conclusion, value.WorkflowName, value.SynthesisJobID, value.SynthesisJobName, value.SynthesisJobStatus, value.SynthesisJobConclusion}},
		{"unique-canonical-artifact", "REGRESSION", "reject-ambiguous-proposal-predecessor", summary.ExactJobs == 1 && summary.ValidCandidates == 1 && summary.AmbiguousCandidates == 0 && summary.UnresolvedCandidates == 0, []int{summary.ExactRuns, summary.ExactJobs, summary.ExactArtifacts, summary.ValidCandidates, summary.AmbiguousCandidates, summary.UnresolvedCandidates}},
		{"ready-proposal-contract", "COHERENCE", "verify-merged-proposal-contract", selected != nil && value.ContractSatisfied == 8 && value.ContractTotal == 8 && value.ContractBPS == 10000 && value.ContractUnresolved == 0, []any{value.ProposalFileSHA256, value.ProposalReportDigest, value.ContractSatisfied, value.ContractTotal, value.ContractBPS}},
		{"read-only-non-authorizing", "FOUNDATION", "preserve-read-only-proposal-selection", selected != nil && value.RepositoryWrites == 0 && !value.PromotionAuthorized && summary.RepositoryWrites == 0, []any{value.RepositoryWrites, value.PromotionAuthorized, summary.RepositoryWrites}},
	}
	result := make([]Proof, 0, len(specs))
	for _, spec := range specs {
		proof, err := makeProof(spec.id, spec.choice, spec.operation, spec.passed, spec.evidence)
		if err != nil {
			return nil, err
		}
		result = append(result, proof)
	}
	return result, nil
}

func canonicalSelected(selected Selected) bool {
	terminal := selected.Conclusion == "success" || selected.Conclusion == "failure"
	return selected.RunID > 0 && selected.RunAttempt > 0 && selected.Event == "push" &&
		selected.Status == "completed" && terminal && selected.WorkflowName == workflowName &&
		selected.SynthesisJobID > 0 && selected.SynthesisJobName == synthesisJobName &&
		selected.SynthesisJobStatus == "completed" && selected.SynthesisJobConclusion == "success" &&
		selected.ArtifactID > 0 && selected.ArtifactName == "metric-strategy-"+selected.HeadSHA
}
