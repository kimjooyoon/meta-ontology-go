package semanticresolution

import (
	"fmt"
	"reflect"
)

func ValidateLatticeReceipt(receipt LatticeReceipt) error {
	if err := validateReceiptIdentity(receipt); err != nil {
		return err
	}
	if err := validateLatticeCases(receipt.Cases); err != nil {
		return err
	}
	if err := validateClaims(receipt.Claims, receipt.Cases); err != nil {
		return err
	}
	if err := validateClaimTamperRegression(receipt.TamperRegression); err != nil {
		return err
	}
	if err := validateCounterfactuals(receipt.Counterfactuals); err != nil {
		return err
	}
	return validateMetrics(receipt.Metrics)
}

func validateLatticeCases(cases []LatticeCase) error {
	for _, item := range cases {
		if item.ID == "" || item.ClaimID == "" || item.Transition.FromResolution != ResolutionExactOperation {
			return fmt.Errorf("lattice case identity is invalid")
		}
		if item.Decision == DecisionUnknown {
			unknown := item.Transition.Unknown
			if item.Transition.Decision != DecisionLowerResolution || item.Transition.ToResolution != ResolutionInvariantOnly || unknown == nil || unknown.Stage != StagePartialObservation || unknown.Step != 1 || unknown.Reason == "" {
				return fmt.Errorf("unknown case did not carry a deterministic descent")
			}
		}
		if item.Decision == DecisionPass && item.Transition.Decision != DecisionPass {
			return fmt.Errorf("pass case is not exact")
		}
		if item.Decision == DecisionFailClosed && item.Transition.Decision != DecisionFailClosed {
			return fmt.Errorf("fail-closed case has an open transition")
		}
		if !reflect.DeepEqual(ReplayPartialObservation(item.Observation), item.Transition) {
			return fmt.Errorf("lattice transition is not replayable")
		}
	}
	return nil
}
