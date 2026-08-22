package proposalpromotion

func sourceFrom(selection selectionView, contract contractView, contractRaw []byte) Source {
	return Source{
		Selection: SelectionSource{
			Repository:        selection.Repository,
			CurrentSubjectSHA: selection.CurrentSubjectSHA,
			PredecessorSHA:    selection.PredecessorSHA,
			Decision:          selection.Decision, Reason: selection.Reason,
			ReportDigest: selection.ReportDigest,
			RunID:        selection.Selected.RunID, RunAttempt: selection.Selected.RunAttempt,
			HeadSHA: selection.Selected.HeadSHA, Event: selection.Selected.Event,
			Status: selection.Selected.Status, Conclusion: selection.Selected.Conclusion,
			WorkflowName: selection.Selected.WorkflowName,
			ArtifactID:   selection.Selected.ArtifactID, ArtifactName: selection.Selected.ArtifactName,
			ProposalFileSHA256:   selection.Selected.ProposalFileSHA256,
			ProposalReportDigest: selection.Selected.ProposalReportDigest,
			ObservedRuns:         selection.Summary.ObservedRuns, ExactRuns: selection.Summary.ExactRuns,
			ObservedArtifacts:           selection.Summary.ObservedArtifacts,
			ExactArtifacts:              selection.Summary.ExactArtifacts,
			ValidCandidates:             selection.Summary.ValidCandidates,
			AmbiguousCandidates:         selection.Summary.AmbiguousCandidates,
			UnresolvedCandidates:        selection.Summary.UnresolvedCandidates,
			SelectionBPS:                selection.Summary.SelectionBPS,
			ProofsPassed:                selection.Summary.ProofsPassed,
			ProofsTotal:                 selection.Summary.ProofsTotal,
			RepositoryWrites:            selection.Summary.RepositoryWrites,
			SelectedRepositoryWrites:    selection.Selected.RepositoryWrites,
			SelectedPromotionAuthorized: selection.Selected.PromotionAuthorized,
		},
		Contract: ContractSource{
			SubjectSHA: contract.SubjectSHA, Decision: contract.Decision, Reason: contract.Reason,
			FileSHA256: digestBytes(contractRaw), ReportDigest: contract.ReportDigest,
			SelectedActions: contract.SelectedActions,
			Satisfied:       contract.Summary.Satisfied, Total: contract.Summary.Total,
			Unresolved: contract.Summary.Unresolved, ReadinessBPS: contract.Summary.ReadinessBPS,
			RepositoryWrites:    contract.RepositoryWrites,
			PromotionAuthorized: contract.PromotionAuthorized,
		},
	}
}
