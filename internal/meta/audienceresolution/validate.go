package audienceresolution

import (
	"fmt"
	"strings"
)

func ValidateReceipt(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.ContractID == "" || receipt.Subject == "" {
		return fmt.Errorf("receipt identity is invalid")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("receipt decision is unknown")
	}
	if (receipt.Resolution != "EXACT" && receipt.Resolution != "LOWER_RESOLUTION" && receipt.Resolution != "INVARIANT_ONLY") ||
		receipt.Reason == "" || !validDigest(receipt.Digest) || !validDigest(receipt.FactsDigest) {
		return fmt.Errorf("receipt resolution or digest is invalid")
	}
	copy := receipt
	copy.Digest = ""
	if digestJSON(copy) != receipt.Digest {
		return fmt.Errorf("receipt digest mismatch")
	}
	if len(receipt.Indicators) != IndicatorTotal || len(receipt.Views) != 3 || len(receipt.Counterexamples) != 2 || len(receipt.ClaimTransitions) != IndicatorTotal {
		return fmt.Errorf("receipt fixed cardinalities are invalid")
	}
	if err := validateIndicators(receipt.Indicators); err != nil {
		return err
	}
	satisfied := 0
	for _, indicator := range receipt.Indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	if receipt.Summary.Coordinates.Total != IndicatorTotal || receipt.Summary.Coordinates.Satisfied != satisfied ||
		receipt.Summary.Coordinates.BasisPoints != basisPoints(satisfied, IndicatorTotal) ||
		receipt.Summary.CounterexamplesBlocked != blockedCounterexamples(receipt.Counterexamples) {
		return fmt.Errorf("receipt summary does not match its indicators")
	}
	if err := validateViews(receipt); err != nil {
		return err
	}
	if !counterexamplesValid(receipt.Counterexamples) {
		return fmt.Errorf("counterexamples are not the fixed blocked cases")
	}
	for index, transition := range receipt.ClaimTransitions {
		indicator := receipt.Indicators[index]
		if transition.IndicatorID != indicator.ID || transition.Before != indicator.ClaimBefore ||
			transition.After != indicator.ClaimAfter || transition.Reason != indicator.Reason {
			return fmt.Errorf("claim transition %q does not match its indicator", transition.IndicatorID)
		}
	}
	if receipt.Decision == "PASS" && (receipt.Resolution != "EXACT" || receipt.Summary.Coordinates.Satisfied != IndicatorTotal) {
		return fmt.Errorf("PASS receipt is not exact")
	}
	for _, view := range receipt.Views {
		if view.Decision != receipt.Decision || view.Reason != receipt.Reason {
			return fmt.Errorf("audience %q contradicts the global decision", view.Audience)
		}
	}
	return nil
}

func validateIndicators(values []Indicator) error {
	specs := indicatorSpecs()
	for index, value := range values {
		spec := specs[index]
		if value.ID != spec.ID || value.Class != spec.Class || value.Producer != spec.Producer ||
			value.Consumer != spec.Consumer || value.MetaOperation != spec.MetaOperation ||
			value.ProofChoice != spec.ProofChoice || value.Stage != spec.Stage || value.Step != spec.Step ||
			value.Reason != spec.Reason || value.ClaimBefore != "UNPROVEN" ||
			(value.Satisfied && value.ClaimAfter != "OBSERVED") || (!value.Satisfied && value.ClaimAfter != "BLOCKED") ||
			value.Expected != 1 || (value.Satisfied && value.Observed != 1) || (!value.Satisfied && value.Observed != 0) {
			return fmt.Errorf("indicator %q is not canonical", value.ID)
		}
	}
	return nil
}

func validateViews(receipt Receipt) error {
	contract := CanonicalContract()
	for index, view := range receipt.Views {
		want := contract.Audiences[index]
		if view.Audience != want.Audience || view.Resolution != want.Resolution ||
			!sameStrings(view.CoordinateIDs, want.Coordinates) || view.Total != len(want.Coordinates) ||
			view.OmittedCoordinateCount != IndicatorTotal-len(want.Coordinates) ||
			view.Satisfied < 0 || view.Satisfied > view.Total || view.BasisPoints != basisPoints(view.Satisfied, view.Total) {
			return fmt.Errorf("audience view %q is not canonical", view.Audience)
		}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, character := range value[7:] {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
