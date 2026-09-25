package toolchainlsp

import (
	"fmt"
	"reflect"
)

var expectedCodeBindings = []string{
	"internal/lsp",
	"internal/lsp/coupling",
	"internal/meta/languagereadiness/toolchainlsp",
	"cmd/toolchain-lsp-witness",
	"examples/toolchain-lsp",
}

func validateConcept(binding ConceptBinding) (string, error) {
	switch binding.ArtifactDecision {
	case DecisionPass:
	case DecisionFailClosed:
		return "TOOLCHAIN_LSP_CONCEPT_FAIL_CLOSED", fmt.Errorf("concept artifact failed closed")
	default:
		return "TOOLCHAIN_LSP_DECISION_UNKNOWN", fmt.Errorf("unknown concept decision %q", binding.ArtifactDecision)
	}
	if binding.ConceptID != "toolchain-lsp" || binding.MetaOperation != MetaOperation || binding.Stage != "OPERATING" {
		return "TOOLCHAIN_LSP_CONCEPT_DRIFT", fmt.Errorf("concept identity drift")
	}
	if !reflect.DeepEqual(binding.CodeBindings, expectedCodeBindings) || !reflect.DeepEqual(binding.MetricBindings, MetricIDs()) || binding.UseCaseBindings != 3 {
		return "TOOLCHAIN_LSP_CONCEPT_DRIFT", fmt.Errorf("concept bindings drift")
	}
	if len(binding.ArtifactDigest) != 71 || binding.ArtifactDigest[:7] != "sha256:" {
		return "TOOLCHAIN_LSP_CONCEPT_DRIFT", fmt.Errorf("concept digest is invalid")
	}
	return "", nil
}
