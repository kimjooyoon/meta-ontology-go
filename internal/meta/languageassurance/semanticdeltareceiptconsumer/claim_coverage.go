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
	byID := make(map[string]Claim, len(ledger))
	for _, claim := range ledger {
		byID[claim.ID] = claim
	}
	rows := []string{"gooo://semantic-delta/claim-transition-identity/v2"}
	for _, transition := range transitions {
		targetSemanticDigest := ""
		if claim, ok := byID[transition.ClaimID]; ok {
			targetSemanticDigest = claim.AfterSemanticDigest
			if targetSemanticDigest == "" {
				targetSemanticDigest = claim.BeforeSemanticDigest
			}
		}
		rows = append(rows, strings.Join([]string{transition.ClaimID, transition.FromStatus, transition.ToStatus, transition.Stage, transition.Step, transition.Reason, targetSemanticDigest}, "\x00"))
	}
	sort.Strings(rows[1:])
	return digestValue(strings.Join(rows, "\x01"))
}

func claimIDInventory(ledger []Claim) []string {
	ids := make([]string, 0, len(ledger))
	for _, claim := range ledger {
		ids = append(ids, claim.ID)
	}
	sort.Strings(ids)
	return ids
}
