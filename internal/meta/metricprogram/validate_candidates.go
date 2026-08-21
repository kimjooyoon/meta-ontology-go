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
	if selection.ProofChoice != "REGRESSION" || selection.Decision != "HOLD_FIXED_POINT" || selection.MetaOperation != "terminate-at-fixed-point" || selection.Reason != "ALL_INDICATORS_SATISFIED_AND_RESIDUALS_ZERO" {
		return fmt.Errorf("only the verified fixed-point strategy is compilable")
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
