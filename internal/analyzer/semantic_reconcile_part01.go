package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// ReconcileWriteEffect is explicit because reconciliation is read-only. A
// rejected identity, binding, or observation can never authorize a write.
type ReconcileWriteEffect string

const ReconcileNoWrite ReconcileWriteEffect = "no-write"

// SemanticReconcileResult is a read-only comparison of an adapted source view
// with an independently declared semantic contract.
type SemanticReconcileResult struct {
	Comparison       semantic.IRComparison
	DeltaValid       bool
	AuthoritySafe    bool
	SourceMatch      bool
	PolicyMatch      bool
	ToolchainMatch   bool
	RegistryMatch    bool
	ObservationMatch bool
	BindingMatch     bool
	Accepted         bool
	WriteEffect      ReconcileWriteEffect
	FailureCode      string
}

// ReconcileSemantic never mutates either IR. Acceptance requires semantic
// equality and all source/policy/toolchain bindings plus provenance equality;
// absent external provenance therefore remains fail-closed.
func ReconcileSemantic(
	observed SemanticAdapterResult, expected semantic.IR, sourceDigest, policyDigest,
	toolchainDigest, observationDigest string,
) SemanticReconcileResult {
	return reconcileSemantic(observed, expected, sourceDigest, policyDigest, toolchainDigest,
		observed.RegistryDigest, observationDigest)
}

// ReconcileSemanticWithRegistry compares a result against an external
// registry digest. Use it when the expected contract names its registry.
func ReconcileSemanticWithRegistry(
	observed SemanticAdapterResult, expected semantic.IR, sourceDigest, policyDigest,
	toolchainDigest, registryDigest, observationDigest string,
) SemanticReconcileResult {
	return reconcileSemantic(observed, expected, sourceDigest, policyDigest, toolchainDigest,
		registryDigest, observationDigest)
}
