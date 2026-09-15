package verify

import (
	"bytes"
	"fmt"
	"slices"
)

func validateInputs(strategy strategyPlan, verification strategyVerification, source []byte) error {
	if strategy.Schema != strategySchema || verification.Schema != strategyVerificationSchema {
		return fmt.Errorf("strategy schema mismatch")
	}
	if strategy.ExecutionPolicy != "READ_ONLY_ARTIFACT_SYNTHESIS" || strategy.RepositoryWorkspaceWrites || strategy.PromotionAuthorized {
		return fmt.Errorf("strategy is not read-only")
	}
	if strategy.Digest != verification.PlanDigest || !validDigest(strategy.Digest) || !validDigest(verification.Digest) || verification.Status != "VERIFIED" {
		return fmt.Errorf("strategy verification binding is invalid")
	}
	if verification.RepositoryWorkspaceWrites || verification.PromotionAuthorized || verification.BindingCount != len(strategy.Bindings) || verification.CandidateCount != len(strategy.Candidates) {
		return fmt.Errorf("strategy verification boundary is invalid")
	}
	canonicalRoot := rootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE", TopologyReason: "ROOT_TOPOLOGY_EXEMPT", ReadmeRequirement: "NOT_APPLICABLE"}
	if strategy.RootPolicy != canonicalRoot {
		return fmt.Errorf("project root exception is not canonical")
	}
	selection := strategy.Selection
	if selection.ProofChoice != "REGRESSION" {
		return fmt.Errorf("strategy is not a regression terminal")
	}
	if selection.MetaOperation == "terminate-at-fixed-point" {
		if selection.Decision != "HOLD_FIXED_POINT" || selection.Reason != "ALL_INDICATORS_SATISFIED_AND_RESIDUALS_ZERO" {
			return fmt.Errorf("strategy is not the verified fixed point")
		}
	} else if selection.MetaOperation == "preserve-non-promoting-terminal" {
		if selection.Decision != "PRESERVE_NON_PROMOTING_TERMINAL" || selection.Reason != "NON_PROMOTING_TERMINAL_PRESERVED" {
			return fmt.Errorf("strategy is not the preserved non-promoting terminal")
		}
	} else {
		return fmt.Errorf("strategy terminal operation is unsupported")
	}
	if verification.SelectedProofChoice != strategy.Selection.ProofChoice {
		return fmt.Errorf("selected proof choice is not verification-bound")
	}
	if !bytes.Equal(source, expectedSource()) {
		return fmt.Errorf("Gooo meta program source is not canonical")
	}
	return validateCandidateBinding(strategy)
}

func validateCandidateBinding(strategy strategyPlan) error {
	if strategy.Input.IndicatorCount != len(strategy.Bindings) {
		return fmt.Errorf("indicator count is not bound")
	}
	for _, binding := range strategy.Bindings {
		if binding.IndicatorID == "" || binding.Status != "SATISFIED" || !validDigest(binding.EvidenceDigest) {
			return fmt.Errorf("indicator %q is not satisfied evidence", binding.IndicatorID)
		}
		operation, ok := findOperation(binding.MetaOperation)
		if !ok || operation.ProofChoice != binding.Family {
			return fmt.Errorf("indicator %q has no independent meta code", binding.IndicatorID)
		}
	}
	for _, candidate := range strategy.Candidates {
		if candidate.ProofChoice != strategy.Selection.ProofChoice {
			continue
		}
		if !candidate.Admissible || candidate.UnsatisfiedCount != 0 || candidate.EvidenceDigest != strategy.Selection.CandidateDigest || !slices.Equal(candidate.MetaOperations, strategy.Selection.SourceMetaOperations) {
			return fmt.Errorf("selected candidate is not canonical")
		}
		return nil
	}
	return fmt.Errorf("selected candidate is missing")
}
