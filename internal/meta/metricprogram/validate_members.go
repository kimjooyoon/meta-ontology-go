package metricprogram

import (
	"fmt"
	"slices"
	"sort"
)

func validateStrategyMembers(plan StrategyPlan, verification StrategyVerification) error {
	if !slices.IsSortedFunc(plan.Bindings, func(left, right StrategyBinding) int { return compareString(left.IndicatorID, right.IndicatorID) }) {
		return fmt.Errorf("strategy bindings are not canonical")
	}
	bindings := make(map[string]StrategyBinding, len(plan.Bindings))
	for _, binding := range plan.Bindings {
		operation, ok := findOperation(binding.MetaOperation)
		if binding.IndicatorID == "" || !ok || binding.Status != "SATISFIED" || !validDigest(binding.EvidenceDigest) {
			return fmt.Errorf("indicator %q is not resolved", binding.IndicatorID)
		}
		if _, duplicate := bindings[binding.IndicatorID]; duplicate {
			return fmt.Errorf("duplicate indicator %q", binding.IndicatorID)
		}
		if operation.ProofChoice != binding.Family || expectedTrilemma(binding.Family) != binding.Trilemma {
			return fmt.Errorf("indicator %q has an invalid proof binding", binding.IndicatorID)
		}
		bindings[binding.IndicatorID] = binding
	}
	return validateCandidates(plan, verification, bindings)
}

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

func expectedTrilemma(choice string) string {
	switch choice {
	case "FOUNDATION":
		return "AXIOM"
	case "COHERENCE":
		return "COHERENCE"
	case "REGRESSION":
		return "REGRESS"
	default:
		return ""
	}
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
