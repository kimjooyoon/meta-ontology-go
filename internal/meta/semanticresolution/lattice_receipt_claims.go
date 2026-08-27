package semanticresolution

import "fmt"

func validateClaims(claims []ClaimRecord, cases []LatticeCase) error {
	if len(claims) != LatticeCaseDenominator {
		return fmt.Errorf("claim ledger denominator is not fixed")
	}
	byClaimID := make(map[string]LatticeCase, len(cases))
	for _, item := range cases {
		if item.ClaimID == "" || byClaimID[item.ClaimID].ID != "" {
			return fmt.Errorf("claim identities are not unique")
		}
		byClaimID[item.ClaimID] = item
	}
	seen := make(map[string]bool, len(claims))
	for _, item := range claims {
		caseItem, ok := byClaimID[item.ID]
		if seen[item.ID] || !ok || item.ID == "" || !validClaimState(item.State) || !validClaimState(item.BeforeState) || !validClaimState(item.AfterState) || item.State != item.AfterState || item.Preserved != (item.BeforeState == item.AfterState) {
			return fmt.Errorf("claim identity or state is invalid")
		}
		if want := deriveClaimAfterState(item.BeforeState, caseItem.Transition.Decision); want != item.AfterState {
			return fmt.Errorf("claim after state was not evidence-derived")
		}
		wantStage, wantStep, wantReason := claimEvidenceFields(caseItem.Transition)
		if item.Stage != wantStage || item.Step != wantStep || item.Reason != wantReason || item.EvidenceDigest != claimEvidenceDigest(item.ID, item.BeforeState, item.AfterState, caseItem.Observation, caseItem.Transition) || item.Provenance != "gooo://semantic-resolution-lattice/case/"+caseItem.ID {
			return fmt.Errorf("claim transition evidence is incomplete")
		}
		seen[item.ID] = true
	}
	return nil
}

func validClaimState(state ClaimState) bool {
	return state == ClaimOpen || state == ClaimDischarged || state == ClaimRefuted
}
