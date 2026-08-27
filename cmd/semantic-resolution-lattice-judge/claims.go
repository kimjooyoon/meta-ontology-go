package main

import "errors"

func validateClaims(claims []claim, cases []latticeCase) error {
	if len(claims) != 4 {
		return errors.New("claim denominator is not fixed")
	}
	seen := map[string]bool{}
	for _, item := range claims {
		if seen[item.ID] || item.ID == "" || !validClaimState(item.State) || item.State != item.AfterState || item.Preserved != (item.BeforeState == item.AfterState) {
			return errors.New("claim state was not preserved")
		}
		seen[item.ID] = true
	}
	for _, item := range cases {
		if !seen[item.ClaimID] {
			return errors.New("case claim is not in the preserved ledger")
		}
	}
	if len(seen) != 4 {
		return errors.New("claim identities are not unique")
	}
	return nil
}

func validClaimState(state string) bool {
	return state == "OPEN" || state == "DISCHARGED" || state == "REFUTED"
}
