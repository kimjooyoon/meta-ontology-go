package analyzer

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// SemanticReconcileResult is a read-only comparison of an adapted source view
// with an independently declared semantic contract.
type SemanticReconcileResult struct {
	Comparison       semantic.IRComparison
	SourceMatch      bool
	PolicyMatch      bool
	ToolchainMatch   bool
	ObservationMatch bool
	Accepted         bool
}

// ReconcileSemantic never mutates either IR. Acceptance requires semantic
// equality and all source/policy/toolchain bindings plus provenance equality;
// absent external provenance therefore remains fail-closed.
func ReconcileSemantic(
	observed SemanticAdapterResult, expected semantic.IR, sourceDigest, policyDigest,
	toolchainDigest, observationDigest string,
) SemanticReconcileResult {
	comparison := semantic.CompareIR(observed.IR, expected)
	result := SemanticReconcileResult{
		Comparison:       comparison,
		SourceMatch:      observed.SourceDigest == sourceDigest,
		PolicyMatch:      observed.PolicyDigest == policyDigest,
		ToolchainMatch:   observed.ToolchainDigest == toolchainDigest,
		ObservationMatch: observed.ImplementationObservationDigest == observationDigest && observationDigest != "",
	}
	result.Accepted = comparison.SemanticEqual && comparison.ProvenanceEqual &&
		result.SourceMatch && result.PolicyMatch && result.ToolchainMatch && result.ObservationMatch
	return result
}
