package main

import (
	"errors"
	"fmt"
)

func validateClaims(claims []claim, cases []latticeCase) error {
	if len(claims) != 4 {
		return errors.New("claim denominator is not fixed")
	}
	seen := map[string]bool{}
	for _, item := range claims {
		caseItem, ok := caseByClaimID(cases, item.ID)
		if seen[item.ID] || !ok || item.ID == "" || !validClaimState(item.State) || !validClaimState(item.BeforeState) || !validClaimState(item.AfterState) || item.State != item.AfterState || item.Preserved != (item.BeforeState == item.AfterState) {
			return errors.New("claim identity or state is invalid")
		}
		if want := deriveClaimAfterState(item.BeforeState, caseItem.Transition.Decision); want != item.AfterState {
			return errors.New("claim after state was not evidence-derived")
		}
		wantStage, wantStep, wantReason := claimEvidenceFields(caseItem.Transition)
		if item.Stage != wantStage || item.Step != wantStep || item.Reason != wantReason || item.EvidenceDigest != claimEvidenceDigest(item.ID, item.BeforeState, item.AfterState, caseItem.Observation, caseItem.Transition) || item.Provenance != "gooo://semantic-resolution-lattice/case/"+caseItem.ID {
			return errors.New("claim transition evidence is incomplete")
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

func caseByClaimID(cases []latticeCase, claimID string) (latticeCase, bool) {
	for _, item := range cases {
		if item.ClaimID == claimID {
			return item, true
		}
	}
	return latticeCase{}, false
}

func validClaimState(state string) bool {
	return state == "OPEN" || state == "DISCHARGED" || state == "REFUTED"
}
