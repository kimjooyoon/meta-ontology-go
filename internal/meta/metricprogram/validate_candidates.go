package metricprogram

import (
	"fmt"
	"slices"
	"sort"
)

func validateCandidates(plan StrategyPlan, verification StrategyVerification, bindings map[string]StrategyBinding) error {
	if len(plan.Candidates) != len(proofChoices) {
		return fmt.Errorf("strategy candidate partition is incomplete")
	}
	covered := make(map[string]bool, len(bindings))
	for index, candidate := range plan.Candidates {
		if candidate.ProofChoice != proofChoices[index] || candidate.IndicatorCount != len(candidate.IndicatorIDs) || candidate.UnsatisfiedCount != 0 || !candidate.Admissible || !validDigest(candidate.EvidenceDigest) {
			return fmt.Errorf("candidate %q is not admissible", candidate.ProofChoice)
		}
		if !slices.IsSorted(candidate.IndicatorIDs) || !slices.IsSorted(candidate.MetaOperations) {
			return fmt.Errorf("candidate %q is not canonical", candidate.ProofChoice)
		}
		operations := make(map[string]bool)
		for _, indicatorID := range candidate.IndicatorIDs {
			binding, ok := bindings[indicatorID]
			if !ok || covered[indicatorID] || binding.Family != candidate.ProofChoice {
				return fmt.Errorf("candidate %q has an invalid indicator partition", candidate.ProofChoice)
			}
			covered[indicatorID] = true
			operations[binding.MetaOperation] = true
		}
		derived := make([]string, 0, len(operations))
		for operation := range operations {
			derived = append(derived, operation)
		}
		sort.Strings(derived)
		if !slices.Equal(derived, candidate.MetaOperations) {
			return fmt.Errorf("candidate %q meta operations are not indicator-derived", candidate.ProofChoice)
		}
	}
	if len(covered) != len(bindings) {
		return fmt.Errorf("candidate partition does not cover every indicator")
	}
	return validateSelection(plan, verification)
}

func validateSelection(plan StrategyPlan, verification StrategyVerification) error {
	selection := plan.Selection
	if selection.ProofChoice != "REGRESSION" {
		return fmt.Errorf("only regression terminal strategies are compilable")
	}
	if selection.MetaOperation == "terminate-at-fixed-point" {
		if selection.Decision != "HOLD_FIXED_POINT" || selection.Reason != "ALL_INDICATORS_SATISFIED_AND_RESIDUALS_ZERO" {
			return fmt.Errorf("fixed-point selection is not canonical")
		}
	} else if selection.MetaOperation == "preserve-non-promoting-terminal" {
		if selection.Decision != "PRESERVE_NON_PROMOTING_TERMINAL" || selection.Reason != "NON_PROMOTING_TERMINAL_PRESERVED" {
			return fmt.Errorf("non-promoting selection is not canonical")
		}
	} else {
		return fmt.Errorf("selected regression operation is unsupported")
	}
	if verification.SelectedProofChoice != selection.ProofChoice || !validDigest(selection.CandidateDigest) {
		return fmt.Errorf("selected proof is not verification-bound")
	}
	for _, candidate := range plan.Candidates {
		if candidate.ProofChoice == selection.ProofChoice {
			if candidate.EvidenceDigest != selection.CandidateDigest || !slices.Equal(candidate.MetaOperations, selection.SourceMetaOperations) {
				return fmt.Errorf("selected candidate binding is invalid")
			}
			return nil
		}
	}
	return fmt.Errorf("selected candidate is missing")
}
