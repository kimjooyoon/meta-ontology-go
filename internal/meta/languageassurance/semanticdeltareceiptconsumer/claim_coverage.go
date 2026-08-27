package semanticdeltareceiptconsumer

import (
	"sort"
	"strings"
)

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
	return explained, len(ledger), coverageBPS(explained, len(ledger))
}

func coverageBPS(numerator, denominator int) int {
	if denominator <= 0 {
		return 0
	}
	return numerator * 10000 / denominator
}

func claimTransitionIdentityDigest(ledger []Claim, transitions []ClaimTransition) string {
	parts := []string{"gooo://semantic-delta/claim-transition-identity/v1"}
	for _, claim := range ledger {
		parts = append(parts, claim.ID, claim.ClaimTypeID, claim.Kind, claim.PropositionDigest, claim.Status, claim.PreservationOf)
	}
	for _, transition := range transitions {
		parts = append(parts, transition.ClaimID, transition.Kind, transition.FromStatus, transition.ToStatus, transition.PropositionDigest, transition.EventID, transition.TransitionDigest, transition.PreviousEventDigest)
	}
	return digestValue(strings.Join(parts, "\x00"))
}

func claimIDInventory(ledger []Claim) []string {
	ids := make([]string, 0, len(ledger))
	for _, claim := range ledger {
		ids = append(ids, claim.ID)
	}
	sort.Strings(ids)
	return ids
}
