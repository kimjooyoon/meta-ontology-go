package claimledger

import (
	"encoding/json"
	"fmt"
)

func validateContract(contract Contract, subject string) error {
	if contract.Schema != ContractSchema {
		return fmt.Errorf("contract schema %q is not %q", contract.Schema, ContractSchema)
	}
	if contract.Metric == "" || subject == "" {
		return fmt.Errorf("metric and subject are required")
	}
	if contract.Expected.FixedClaimTotal != len(contract.Claims) {
		return fmt.Errorf("fixed claim total %d does not match %d claims", contract.Expected.FixedClaimTotal, len(contract.Claims))
	}
	seen := map[string]bool{}
	for _, spec := range contract.Claims {
		if err := validateClaim(spec, seen); err != nil {
			return err
		}
		seen[spec.ID] = true
	}
	return nil
}

func validateClaim(spec ClaimSpec, seen map[string]bool) error {
	if spec.ID == "" || seen[spec.ID] {
		return fmt.Errorf("claim id %q is empty or duplicated", spec.ID)
	}
	if spec.Kind == "" || spec.Modality == "" || spec.Subject == "" || spec.Predicate == "" {
		return fmt.Errorf("claim %q is incomplete", spec.ID)
	}
	if !validStage(spec.Coordinate.Stage) || spec.Coordinate.Step == "" {
		return fmt.Errorf("claim %q has no precise process coordinate", spec.ID)
	}
	if !validProofRoute(spec.ProofRoute) {
		return fmt.Errorf("claim %q has invalid proof route %q", spec.ID, spec.ProofRoute)
	}
	if spec.Scope == "EXCLUDED" {
		if spec.ExcludedReason == "" {
			return fmt.Errorf("excluded claim %q has no reason", spec.ID)
		}
		return nil
	}
	if spec.Scope != "IN_SCOPE" || spec.Evidence == nil || spec.Evidence.Source == "" || len(spec.Evidence.Paths) == 0 {
		return fmt.Errorf("in-scope claim %q has no evidence selector", spec.ID)
	}
	if !validOperator(spec.Evidence.Operator) {
		return fmt.Errorf("claim %q has invalid operator %q", spec.ID, spec.Evidence.Operator)
	}
	if spec.Evidence.Operator == "EQUALS" && !json.Valid(spec.Evidence.Expected) {
		return fmt.Errorf("claim %q has no valid expected value", spec.ID)
	}
	if spec.UnknownReason == "" || spec.RefutedReason == "" {
		return fmt.Errorf("claim %q has incomplete failure reasons", spec.ID)
	}
	return nil
}
