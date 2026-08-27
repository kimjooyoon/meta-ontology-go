package semanticdeltareceipt

func claimStatusCoverage(ledger []Claim, transitions []ClaimTransition) (int, int, int) {
	byClaim := make(map[string][]ClaimTransition, len(transitions))
	for _, transition := range transitions {
		byClaim[transition.ClaimID] = append(byClaim[transition.ClaimID], transition)
	}
	explained := 0
	for _, claim := range ledger {
		matches := byClaim[claim.ID]
		if len(matches) != 1 {
			continue
		}
		transition := matches[0]
		if transition.ToStatus == claim.Status && transition.PropositionDigest == claim.PropositionDigest {
			explained++
		}
	}
	return explained, len(ledger), semanticCoverageBPS(explained, len(ledger))
}
