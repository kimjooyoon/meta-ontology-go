package proposalpromotion

func buildCoordinates(currentHead, evidenceHead string, source Source) []Coordinate {
	selection, contract := source.Selection, source.Contract
	return []Coordinate{
		coordinate("exact-current-subject", "FOUNDATION",
			selection.CurrentSubjectSHA == currentHead && currentHead != evidenceHead,
			[]string{selection.CurrentSubjectSHA, currentHead, evidenceHead}),
		coordinate("exact-merged-predecessor", "FOUNDATION",
			selection.PredecessorSHA == evidenceHead && selection.HeadSHA == evidenceHead &&
				contract.SubjectSHA == evidenceHead,
			[]string{selection.PredecessorSHA, selection.HeadSHA,
				contract.SubjectSHA, evidenceHead}),
		coordinate("canonical-push-synthesis", "COHERENCE",
			selection.Event == "push" && selection.Status == "completed" &&
				selection.RequestedRoute != "" && selection.HeadBranch == selection.RequestedRoute &&
				selection.WorkflowName == "Metric counterfactual conformance" &&
				selection.SynthesisJobID > 0 && selection.SynthesisJobName == "strategy" &&
				selection.SynthesisJobStatus == "completed" &&
				selection.SynthesisJobConclusion == "success" &&
				selection.ArtifactName == "metric-strategy-"+evidenceHead,
			[]any{selection.RequestedRoute, selection.HeadBranch, selection.Event, selection.Status, selection.Conclusion,
				selection.WorkflowName, selection.SynthesisJobID, selection.SynthesisJobName,
				selection.SynthesisJobStatus, selection.SynthesisJobConclusion,
				selection.ArtifactName}),
		coordinate("unique-canonical-artifact", "REGRESSION",
			selection.ExactRuns == 1 && selection.ExactJobs == 1 &&
				selection.ExactArtifacts == 1 &&
				selection.ValidCandidates == 1 && selection.AmbiguousCandidates == 0,
			[]int{selection.ExactRuns, selection.ExactJobs, selection.ExactArtifacts,
				selection.ValidCandidates, selection.AmbiguousCandidates}),
		coordinate("ready-proposal-contract", "COHERENCE",
			contract.Decision == "PASS" && contract.Satisfied == 8 &&
				contract.Reason == "CHANGE_PROPOSAL_CONTRACT_READY" &&
				contract.Total == 8 && contract.ReadinessBPS == 10_000 &&
				selection.ProposalFileSHA256 == contract.FileSHA256 &&
				selection.ProposalReportDigest == contract.ReportDigest,
			[]any{contract.Decision, contract.Reason, contract.Satisfied,
				contract.Total, contract.ReadinessBPS, selection.ProposalFileSHA256,
				contract.FileSHA256, selection.ProposalReportDigest, contract.ReportDigest}),
		coordinate("complete-selection-proofs", "COHERENCE",
			selection.Decision == "SELECTED" &&
				selection.Reason == "PROPOSAL_PREDECESSOR_SELECTED" &&
				selection.SelectionBPS == 10_000 && selection.ProofsPassed == 5 &&
				selection.ProofsTotal == 5,
			[]any{selection.Decision, selection.Reason, selection.SelectionBPS,
				selection.ProofsPassed, selection.ProofsTotal}),
		coordinate("resolved-read-only-evidence", "REGRESSION",
			selection.UnresolvedCandidates == 0 && contract.Unresolved == 0 &&
				selection.RepositoryWrites == 0 && selection.SelectedRepositoryWrites == 0 &&
				contract.RepositoryWrites == 0,
			[]int{selection.UnresolvedCandidates, contract.Unresolved,
				selection.RepositoryWrites, selection.SelectedRepositoryWrites,
				contract.RepositoryWrites}),
		coordinate("non-authorizing-boundary", "FOUNDATION",
			!selection.SelectedPromotionAuthorized && !contract.PromotionAuthorized,
			[]bool{selection.SelectedPromotionAuthorized, contract.PromotionAuthorized}),
	}
}

func coordinate(id, choice string, satisfied bool, evidence any) Coordinate {
	status, reason := "UNRESOLVED", "COORDINATE_EVIDENCE_UNKNOWN"
	if satisfied {
		status, reason = "SATISFIED", "COORDINATE_EXACTLY_PROVEN"
	}
	return Coordinate{
		ID: id, ProofChoice: choice, Status: status, Reason: reason,
		EvidenceDigest: digestJSON(evidence),
	}
}
