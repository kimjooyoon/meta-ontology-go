package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func reconcileSemantic(
	observed SemanticAdapterResult, expected semantic.IR, sourceDigest, policyDigest,
	toolchainDigest, registryDigest, observationDigest string,
) SemanticReconcileResult {
	comparison := semantic.CompareIR(observed.IR, expected)
	deltaValid := normalizedDeltaValid(observed)
	authoritySafe := normalizedDeltaAuthoritySafe(observed)
	result := SemanticReconcileResult{
		Comparison:       comparison,
		DeltaValid:       deltaValid,
		AuthoritySafe:    authoritySafe,
		SourceMatch:      observed.SourceDigest == sourceDigest,
		PolicyMatch:      observed.PolicyDigest == policyDigest,
		ToolchainMatch:   observed.ToolchainDigest == toolchainDigest,
		RegistryMatch:    observed.RegistryDigest == registryDigest && registryDigest != "",
		ObservationMatch: observed.ImplementationObservationDigest == observationDigest && observationDigest != "",
		BindingMatch:     observed.BindingDigest == semanticAdapterBindingDigest(observed),
		WriteEffect:      ReconcileNoWrite,
	}
	result.Accepted = deltaValid && authoritySafe && comparison.SemanticEqual && comparison.ProvenanceEqual &&
		result.SourceMatch && result.PolicyMatch && result.ToolchainMatch && result.RegistryMatch &&
		result.ObservationMatch && result.BindingMatch
	if !result.Accepted {
		result.FailureCode = reconcileFailureCode(result)
	}
	return result
}
