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
		{"canonical-push-workflow", "COHERENCE", "bind-canonical-proposal-workflow", selected != nil && canonicalSelected(value), []any{value.RunID, value.RunAttempt, value.Event, value.Status, value.Conclusion, value.WorkflowName}},
		{"unique-canonical-artifact", "REGRESSION", "reject-ambiguous-proposal-predecessor", summary.ValidCandidates == 1 && summary.AmbiguousCandidates == 0 && summary.UnresolvedCandidates == 0, []int{summary.ExactRuns, summary.ExactArtifacts, summary.ValidCandidates, summary.AmbiguousCandidates, summary.UnresolvedCandidates}},
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
	return selected.RunID > 0 && selected.RunAttempt > 0 && selected.Event == "push" && selected.Status == "completed" && selected.Conclusion == "success" && selected.WorkflowName == workflowName && selected.ArtifactID > 0 && selected.ArtifactName == "metric-strategy-"+selected.HeadSHA
}
