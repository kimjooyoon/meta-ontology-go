package semanticresolution

import "fmt"

func validateClaims(claims []ClaimRecord) error {
	if len(claims) != LatticeCaseDenominator {
		return fmt.Errorf("claim ledger denominator is not fixed")
	}
	for _, item := range claims {
		if item.ID == "" || !validClaimState(item.State) || item.State != item.BeforeState || item.State != item.AfterState || !item.Preserved {
			return fmt.Errorf("claim state was not preserved")
		}
	}
	return nil
}

func validClaimState(state ClaimState) bool {
	return state == ClaimOpen || state == ClaimDischarged || state == ClaimRefuted
}
