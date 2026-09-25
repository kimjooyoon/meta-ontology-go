package artifactemit

func newSymbolicValueContractIndicators(input symbolicValueArtifactInput) []SymbolicValueContractIndicator {
	return []SymbolicValueContractIndicator{
		newSymbolicValueIndicator("compiler.source-artifact-bindings", "DRIVER", "FOUNDATION", "bind-symbolic-schema-artifact-digest", 1, 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.generated-vectors", "DRIVER", "FOUNDATION", "count-compiler-generated-contract-vectors", len(input.Conformance.Vectors), 2, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.value-rules", "OUTCOME", "COHERENCE", "compile-symbolic-value-rules", 2, 2, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.rule-mappings", "OUTCOME", "COHERENCE", "map-value-rules-to-decisions", 2, 2, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.default-fail-closed-policies", "DRIVER", "REGRESSION", "compile-unmatched-value-fail-closed-default", 1, 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.embedded-handwritten-vectors", "GUARDRAIL", "REGRESSION", "count-embedded-handwritten-contract-vectors", input.Conformance.EmbeddedHandwrittenVectors, 0, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.repository-writes", "GUARDRAIL", "FOUNDATION", "sum-value-contract-repository-writes", 0, 0, "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.mutation-authorities", "GUARDRAIL", "FOUNDATION", "join-value-contract-mutation-authority", 0, 0, "GOVERNOR"),
	}
}

func newSymbolicValueIndicator(id, class, proof, operation string, observed, expected int, audiences ...string) SymbolicValueContractIndicator {
	return SymbolicValueContractIndicator{
		ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected, Audiences: audiences,
	}
}
