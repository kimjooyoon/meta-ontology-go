package audienceresolution

import (
	"fmt"
	"strings"
)

// ValidateReceipt is a producer-side shape check only. The CI consumer is the
// independent checker and intentionally re-parses raw source and raw ledger.
func ValidateReceipt(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.ContractID == "" || receipt.Subject == "" {
		return fmt.Errorf("receipt identity is invalid")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "UNKNOWN" && receipt.Decision != "REFUTED" {
		return fmt.Errorf("receipt decision is unknown")
	}
	if receipt.Resolution != "EXACT" && receipt.Resolution != "LOWER_RESOLUTION" && receipt.Resolution != "INVARIANT_ONLY" ||
		receipt.Reason == "" || !validDigest(receipt.Digest) || !validDigest(receipt.FactsDigest) {
		return fmt.Errorf("receipt resolution or digest is invalid")
	}
	copy := receipt
	copy.Digest = ""
	if receiptDigest(copy) != receipt.Digest {
		return fmt.Errorf("receipt digest mismatch")
	}
	if len(receipt.Indicators) != IndicatorTotal || len(receipt.Views) != 3 || len(receipt.Counterexamples) != 2 || len(receipt.ClaimTransitions) != 36 {
		return fmt.Errorf("receipt fixed cardinalities are invalid")
	}
	if err := validateIndicators(receipt.Indicators); err != nil {
		return err
	}
	satisfied := countSatisfied(receipt.Indicators)
	if receipt.Summary.Coordinates.Total != IndicatorTotal || receipt.Summary.Coordinates.Satisfied != satisfied ||
		receipt.Summary.Coordinates.BasisPoints != basisPoints(satisfied, IndicatorTotal) || receipt.Summary.SourceDenominator <= 0 {
		return fmt.Errorf("receipt summary does not match its indicators")
	}
	if err := validateViews(receipt); err != nil {
		return err
	}
	if err := validateCounterexamples(receipt.Counterexamples); err != nil {
		return err
	}
	for _, transition := range receipt.ClaimTransitions {
		if transition.IndicatorID == "" || transition.Audience == "" || transition.Before == "" || transition.After == "" ||
			transition.Visibility == "" || !validDigest(transition.EvidenceDigest) || transition.Producer == "" || transition.Consumer == "" ||
			transition.Stage == "" || transition.Step == "" || transition.Reason == "" {
			return fmt.Errorf("claim transition %q is incomplete", transition.IndicatorID)
		}
		if transition.Before != "OPEN" || (transition.After != "OPEN" && transition.After != "DISCHARGED" && transition.After != "REFUTED") ||
			(transition.Visibility != "VISIBLE" && transition.Visibility != "OMITTED") {
			return fmt.Errorf("claim transition %q is not append-only", transition.IndicatorID)
		}
	}
	return nil
}

func validateIndicators(values []Indicator) error {
	specs := indicatorSpecs()
	for index, value := range values {
		spec := specs[index]
		if value.ID != spec.ID || value.Class != spec.Class || value.Producer != spec.Producer || value.Consumer != spec.Consumer ||
			value.MetaOperation != spec.MetaOperation || value.ProofChoice != spec.ProofChoice || value.Stage != spec.Stage ||
			value.Step != spec.Step || value.Reason != spec.Reason || value.ClaimBefore != "OPEN" || value.Expected != 1 ||
			(value.Satisfied && (value.Observed != 1 || value.ClaimAfter != "DISCHARGED")) ||
			(!value.Satisfied && value.Observed != 0 && value.ClaimAfter != "REFUTED") {
			return fmt.Errorf("indicator %q is not canonical", value.ID)
		}
		if value.ClaimAfter != "OPEN" && value.ClaimAfter != "DISCHARGED" && value.ClaimAfter != "REFUTED" {
			return fmt.Errorf("indicator %q has invalid claim state", value.ID)
		}
	}
	return nil
}

func validateViews(receipt Receipt) error {
	if receipt.Views[0].Audience != "USER" || receipt.Views[1].Audience != "TOOL_AUTHOR" || receipt.Views[2].Audience != "GOVERNOR" {
		return fmt.Errorf("audience order is not canonical")
	}
	governor := receipt.Views[2]
	if governor.Required != len(governor.CoordinateIDs) || governor.Total != governor.Required || governor.Visible != governor.Required {
		return fmt.Errorf("governor view is not complete")
	}
	for index, view := range receipt.Views {
		if view.Resolution == "" || view.Required != governor.Required || view.Total != len(view.CoordinateIDs) ||
			view.Visible < 0 || view.Visible > view.Required || view.OmittedCoordinateCount != view.Required-view.Visible ||
			view.Satisfied < 0 || view.Satisfied > view.Total || view.BasisPoints != basisPoints(view.Satisfied, view.Total) ||
			view.GlobalDecision != receipt.Decision || view.LocalDecision == "" || view.LocalResolution == "" || view.LocalReason == "" ||
			(view.InheritedStatus != "LOCALLY_VERIFIED" && view.InheritedStatus != "INHERITED_NOT_LOCALLY_VERIFIED") {
			return fmt.Errorf("audience view %q is not canonical", view.Audience)
		}
		if index > 0 && !prefix(receipt.Views[index-1].CoordinateIDs, view.CoordinateIDs) {
			return fmt.Errorf("audience coordinate policy is not nested")
		}
	}
	return nil
}

func validateCounterexamples(values []CounterexampleResult) error {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value.ID] || value.ID == "" || value.Global == "PASS" || len(value.Views) != 3 || value.Reason == "" {
			return fmt.Errorf("counterexample %q is not executed", value.ID)
		}
		seen[value.ID] = true
	}
	if !seen["counterexample.missing-information"] || !seen["counterexample.decision-contradiction"] {
		return fmt.Errorf("counterexample execution set is incomplete")
	}
	return nil
}

func prefix(shorter, longer []string) bool {
	if len(longer) < len(shorter) {
		return false
	}
	for index, value := range shorter {
		if value != longer[index] {
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
