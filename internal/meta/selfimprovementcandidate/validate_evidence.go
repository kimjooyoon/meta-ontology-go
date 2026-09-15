package selfimprovementcandidate

func validCandidate(candidate Candidate, sourceDigest string) bool {
	return candidate.Schema == CandidateSchema && validDigest(candidate.ID) &&
		candidate.SourceObservationDigest == sourceDigest && candidate.GapID == "value-level-computation" &&
		candidate.SourceNonClaim == "value-level computation" &&
		candidate.ExperimentKind == "VALUE_WITNESS_EXPERIMENT" &&
		coordinateEquals(candidate.Before, 0, 1) && coordinateEquals(candidate.Target, 1, 1) &&
		candidate.ProofChoice == "COHERENCE" &&
		candidate.MetaOperation == "propose-value-level-witness-experiment" &&
		validDigest(candidate.ExecutionInputDigest) &&
		!candidate.ExecutionAuthorized && !candidate.MutationAuthorized &&
		!candidate.PromotionAuthorized && !candidate.AutomaticAdoptionAuthorized &&
		candidate.Digest == candidateDigest(candidate)
}

func evidenceState(report Report, success bool) bool {
	for _, indicator := range report.Indicators {
		if indicator.ID == "" || indicator.Class == "" || indicator.ProofChoice == "" ||
			indicator.MetaOperation == "" || indicator.Target != 1 ||
			indicator.Value != boolInt(success) || indicator.Satisfied != success {
			return false
		}
	}
	expectedTotals := []int{5, 12, 16}
	for index, view := range report.Views {
		if view.Total != expectedTotals[index] || len(view.IndicatorIDs) != view.Total ||
			view.Satisfied != boolInt(success)*view.Total ||
			view.BasisPoints != boolInt(success)*10_000 {
			return false
		}
	}
	for _, proof := range report.Proofs {
		if proof.Choice == "" || proof.Claim == "" || proof.MetaOperation == "" ||
			!validDigest(proof.EvidenceDigest) || proof.Passed != success {
			return false
		}
	}
	return true
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
