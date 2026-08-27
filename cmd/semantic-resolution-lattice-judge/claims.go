package main

import "errors"

func validateClaims(claims []claim, cases []latticeCase) error {
	if len(claims) != 4 {
		return errors.New("claim denominator is not fixed")
	}
	expected := map[string]string{
		"claim-exact-observation":            "DISCHARGED",
		"claim-invariant-fallback":           "OPEN",
		"claim-exact-under-missing-evidence": "REFUTED",
		"claim-write-free-descent":           "DISCHARGED",
	}
	seen := map[string]bool{}
	for _, item := range claims {
		want, known := expected[item.ID]
		if !known || seen[item.ID] || item.State != want || item.State != item.BeforeState || item.State != item.AfterState || !item.Preserved {
			return errors.New("claim state was not preserved")
		}
		seen[item.ID] = true
	}
	for _, item := range cases {
		if !seen[item.ClaimID] {
			return errors.New("case claim is not in the preserved ledger")
		}
	}
	return nil
}
