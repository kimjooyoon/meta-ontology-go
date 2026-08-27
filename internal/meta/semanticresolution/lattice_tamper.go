package semanticresolution

import (
	"fmt"
	"strings"
)

const (
	claimTamperRegressionID = "legacy-final-state-tamper"
	claimTamperTargetID     = "claim-write-free-descent"
	claimTamperMutation     = "add claim_state=DISCHARGED to mutation-authority value"
)

func buildClaimTamperRegression(sourcePath, source string) (LatticeTamperRegression, error) {
	tampered, err := appendCaseField(source, "mutation-authority", "claim_state=DISCHARGED")
	if err != nil {
		return LatticeTamperRegression{}, err
	}
	_, claims, _, parseErr := casesFromGoooSource(sourcePath, tampered)
	minted := false
	if parseErr == nil {
		for _, claim := range claims {
			if claim.ID == claimTamperTargetID && claim.AfterState == ClaimDischarged {
				minted = true
				break
			}
		}
	}
	return LatticeTamperRegression{ID: claimTamperRegressionID, ClaimID: claimTamperTargetID,
		Mutation: claimTamperMutation, Rejected: parseErr != nil, MintedDischarged: minted}, nil
}

func appendCaseField(source, caseID, field string) (string, error) {
	marker := latticeCaseProgramPrefix + "id=" + caseID + ";"
	start := strings.Index(source, marker)
	if start < 0 {
		return "", fmt.Errorf("case %q is missing from source", caseID)
	}
	lineEnd := strings.IndexByte(source[start:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source) - start
	}
	quote := strings.LastIndex(source[start:start+lineEnd], `"`)
	if quote < 0 {
		return "", fmt.Errorf("case %q has no value terminator", caseID)
	}
	return source[:start+quote] + ";" + field + source[start+quote:], nil
}

func validateClaimTamperRegression(item LatticeTamperRegression) error {
	if item.ID != claimTamperRegressionID || item.ClaimID != claimTamperTargetID || item.Mutation != claimTamperMutation || !item.Rejected || item.MintedDischarged {
		return fmt.Errorf("claim final-state tamper regression did not fail closed")
	}
	return nil
}
