package semanticdeltareceipt

import (
	"sort"
	"strings"
)

func finishReceiptClaims(receipt *Receipt) {
	receipt.TransitionCount = len(receipt.ClaimTransitions)
	if receipt.TransitionCount > 0 {
		receipt.TransitionHeadDigest = receipt.ClaimTransitions[receipt.TransitionCount-1].TransitionDigest
	}
	receipt.ClaimIDInventory = make([]string, 0, len(receipt.ClaimLedger))
	for _, claim := range receipt.ClaimLedger {
		receipt.ClaimIDInventory = append(receipt.ClaimIDInventory, claim.ID)
	}
	sort.Strings(receipt.ClaimIDInventory)
	receipt.ClaimTransitionIdentityDigest = claimTransitionIdentityDigest(receipt.ClaimLedger, receipt.ClaimTransitions)
	receipt.ClaimsWithExplainedStatus, receipt.TotalClaims, receipt.ClaimStatusCoverageBPS = claimStatusCoverage(receipt.ClaimLedger, receipt.ClaimTransitions)
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
