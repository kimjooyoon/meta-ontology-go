package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
)

// ClaimStateExpectation is validator-owned expectation data. Producer
// observation never reads this table: it records the states it actually
// derives, while the verifier compares those observations with this fixed
// external contract.
type ClaimStateExpectation struct {
	CaseID string   `json:"case_id"`
	States []string `json:"states"`
}

type claimStateTotals struct {
	Discharged int
	Open       int
	Refuted    int
}

type claimStateMismatch struct {
	CaseID   string `json:"case_id"`
	ClaimID  string `json:"claim_id"`
	Actual   string `json:"actual"`
	Expected string `json:"expected"`
}

type claimStateExpectationDetail struct {
	FixedDenominator int                  `json:"fixed_denominator"`
	Mismatches       []claimStateMismatch `json:"mismatches"`
	ActualTotals     claimStateTotalsJSON `json:"actual_totals"`
	ExpectedTotals   claimStateTotalsJSON `json:"expected_totals"`
}

type claimStateTotalsJSON struct {
	Discharged int `json:"discharged"`
	Open       int `json:"open"`
	Refuted    int `json:"refuted"`
}

// fixedClaimStateExpectations is intentionally maintained independently from
// Evaluate and summarize. It is the 16-case x 5-claim validator expectation
// table for the v3 fixed denominator (80 claim instances).
var fixedClaimStateExpectations = []ClaimStateExpectation{
	{CaseID: "valid-proof-carrying-artifact", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED"}},
	{CaseID: "tampered-evidence", States: []string{"REFUTED", "OPEN", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "coherent-tamper-reconstruction", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "missing-operation-evidence", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "bytes-only-no-authority", States: []string{"OPEN", "OPEN", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "independent-recipe-mismatch", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "recipe-only-mismatch", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED", "OPEN"}},
	{CaseID: "missing-attachment", States: []string{"DISCHARGED", "OPEN", "DISCHARGED", "OPEN", "OPEN"}},
	{CaseID: "wrong-attachment-digest", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "unrelated-evidence-tamper", States: []string{"DISCHARGED", "DISCHARGED", "REFUTED", "OPEN", "OPEN"}},
	{CaseID: "stale-head", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED"}},
	{CaseID: "unauthorized-consumer", States: []string{"DISCHARGED", "DISCHARGED", "DISCHARGED", "DISCHARGED", "REFUTED"}},
	{CaseID: "coherent-claim-proposition-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-dependency-tamper", States: []string{"DISCHARGED", "REFUTED", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-proof-choice-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
	{CaseID: "coherent-claim-target-tamper", States: []string{"REFUTED", "OPEN", "DISCHARGED", "DISCHARGED", "OPEN"}},
}

func fixedClaimStateTable() []ClaimStateExpectation {
	result := make([]ClaimStateExpectation, len(fixedClaimStateExpectations))
	for index, item := range fixedClaimStateExpectations {
		result[index] = ClaimStateExpectation{CaseID: item.CaseID, States: append([]string(nil), item.States...)}
	}
	return result
}

// ClaimStateExpectations exposes the validator-owned table for CI evidence.
// It is deliberately not used by Evaluate when deriving observed states.
func ClaimStateExpectations() []ClaimStateExpectation {
	return fixedClaimStateTable()
}

func fixedClaimStateTotals() claimStateTotals {
	var totals claimStateTotals
	for _, item := range fixedClaimStateExpectations {
		for _, state := range item.States {
			switch state {
			case "DISCHARGED":
				totals.Discharged++
			case "OPEN":
				totals.Open++
			case "REFUTED":
				totals.Refuted++
			default:
				panic("invalid fixed claim state expectation")
			}
		}
	}
	return totals
}

func fixedClaimStateTotalsJSON() claimStateTotalsJSON {
	totals := fixedClaimStateTotals()
	return claimStateTotalsJSON{Discharged: totals.Discharged, Open: totals.Open, Refuted: totals.Refuted}
}

func observedClaimStateTotals(cases []CaseResult) claimStateTotalsJSON {
	var totals claimStateTotalsJSON
	for _, item := range cases {
		for _, claim := range item.Claims {
			switch claim.Status {
			case "DISCHARGED":
				totals.Discharged++
			case "OPEN":
				totals.Open++
			case "REFUTED":
				totals.Refuted++
			}
		}
	}
	return totals
}

func validateClaimStateExpectations(cases []CaseResult) error {
	expectations := fixedClaimStateExpectations
	if len(expectations) != CaseTotal || len(expectations)*ClaimTemplateTotal != CaseTotal*ClaimTemplateTotal {
		return fmt.Errorf("fixed claim state expectation inventory has denominator %d, want %d", len(expectations)*ClaimTemplateTotal, CaseTotal*ClaimTemplateTotal)
	}
	claimIDs := make([]string, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		claimIDs = append(claimIDs, spec.ID)
	}
	mismatches := make([]claimStateMismatch, 0)
	for caseIndex, expected := range expectations {
		if expected.CaseID != CaseIDs()[caseIndex] || len(expected.States) != ClaimTemplateTotal || cases[caseIndex].ID != expected.CaseID {
			return fmt.Errorf("fixed claim state expectation case inventory mismatch at index %d", caseIndex)
		}
		for claimIndex, expectedState := range expected.States {
			if cases[caseIndex].Claims[claimIndex].ID != claimIDs[claimIndex] {
				return fmt.Errorf("fixed claim state expectation claim inventory mismatch: %s", expected.CaseID)
			}
			if cases[caseIndex].Claims[claimIndex].Status != expectedState {
				mismatches = append(mismatches, claimStateMismatch{CaseID: expected.CaseID, ClaimID: claimIDs[claimIndex], Actual: cases[caseIndex].Claims[claimIndex].Status, Expected: expectedState})
			}
		}
	}
	if len(mismatches) == 0 {
		return nil
	}
	detail, err := json.Marshal(claimStateExpectationDetail{FixedDenominator: CaseTotal * ClaimTemplateTotal, Mismatches: mismatches, ActualTotals: observedClaimStateTotals(cases), ExpectedTotals: fixedClaimStateTotalsJSON()})
	if err != nil {
		return fmt.Errorf("claim state expectation mismatch")
	}
	return &ValidationError{Coordinate: Coordinate{"VERIFY_CLAIM_STATES", "case-claim-state", "CLAIM_STATE_EXPECTATION_MISMATCH"}, Detail: string(detail)}
}
