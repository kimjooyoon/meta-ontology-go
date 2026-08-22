package proposalpredecessor

import "fmt"

func Select(repository, currentSHA, predecessorSHA string, collection Collection) (Report, []byte, error) {
	summary := Summary{ObservedRuns: collection.ObservedRuns, ExactRuns: collection.ExactRuns, ObservedArtifacts: collection.ObservedArtifacts, ExactArtifacts: collection.ExactArtifacts, ValidCandidates: len(collection.Candidates), UnresolvedCandidates: collection.Unresolved, ProofsTotal: 5}
	if len(collection.Candidates) > 1 {
		summary.AmbiguousCandidates = len(collection.Candidates) - 1
	}
	if len(collection.Candidates) == 0 && summary.UnresolvedCandidates == 0 {
		summary.UnresolvedCandidates = 1
	}
	decision, reason := "FAIL_CLOSED", "PROPOSAL_PREDECESSOR_NOT_FOUND"
	var selected *Selected
	var payload []byte
	if len(collection.Candidates) > 1 {
		reason = "PROPOSAL_PREDECESSOR_AMBIGUOUS"
	} else if len(collection.Candidates) == 1 && summary.UnresolvedCandidates != 0 {
		reason = "PROPOSAL_PREDECESSOR_EVIDENCE_UNKNOWN"
	} else if len(collection.Candidates) == 1 && candidateReady(collection.Candidates[0], predecessorSHA) {
		candidate := collection.Candidates[0]
		selected, payload = &candidate.Selected, candidate.ProposalPayload
		decision, reason, summary.SelectionBPS = "SELECTED", "PROPOSAL_PREDECESSOR_SELECTED", 10000
	}
	proofs, err := buildProofs(selected, predecessorSHA, summary)
	if err != nil {
		return Report{}, nil, err
	}
	for _, proof := range proofs {
		if proof.Passed {
			summary.ProofsPassed++
		}
	}
	report := Report{Schema: Schema, Repository: repository, CurrentSubjectSHA: currentSHA, PredecessorSHA: predecessorSHA, Decision: decision, Reason: reason, Selected: selected, Summary: summary, Indicators: buildIndicators(summary), Proofs: proofs}
	report, err = sealReport(report)
	if err != nil {
		return Report{}, nil, err
	}
	if err := Validate(report); err != nil {
		return Report{}, nil, err
	}
	if !report.Ready() {
		return report, nil, fmt.Errorf("proposal predecessor selection failed closed: %s", report.Reason)
	}
	return report, payload, nil
}

func candidateReady(candidate Candidate, predecessorSHA string) bool {
	selected := candidate.Selected
	return selected.HeadSHA == predecessorSHA && selected.ContractSatisfied == 8 && selected.ContractTotal == 8 && selected.ContractBPS == 10000 && selected.ContractUnresolved == 0 && selected.RepositoryWrites == 0 && !selected.PromotionAuthorized && canonicalSelected(selected)
}
