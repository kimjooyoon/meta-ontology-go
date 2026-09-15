package artifactemit

func newSymbolicValueReachabilityIndicators(artifact Artifact, contract SymbolicValueContract, subjectSHA string, analysis symbolicValueReachabilityAnalysis) []SymbolicValueContractIndicator {
	return []SymbolicValueContractIndicator{
		newSymbolicValueIndicator("compiler.source-artifact-bindings", "DRIVER", "FOUNDATION", "bind-reachability-to-symbolic-artifact", boolCount(contract.SourceArtifactDigest == artifact.Digest), 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.source-contract-bindings", "DRIVER", "FOUNDATION", "bind-reachability-to-value-contract", boolCount(contract.SubjectSHA == subjectSHA && validSymbolicValueSHA256(contract.Digest)), 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.supported-schema-profiles", "DRIVER", "FOUNDATION", "recognize-generated-symbolic-schema-profile", boolCount(analysis.SchemaProfileSupported), 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.schema-ready-entailments", "OUTCOME", "COHERENCE", "compile-schema-to-ready-rule-entailment", boolCount(analysis.SchemaEntailsReady), 1, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.reachable-value-rules", "OUTCOME", "COHERENCE", "count-rules-reachable-after-structural-gate", analysis.Summary.ReachableRules, 1, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.defense-only-value-rules", "OUTCOME", "REGRESSION", "count-rules-unreachable-after-structural-gate", analysis.Summary.DefenseOnlyRules, 1, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.reachable-default-policies", "GUARDRAIL", "REGRESSION", "count-defaults-reachable-after-structural-gate", analysis.Summary.ReachableDefaults, 0, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.defense-only-default-policies", "OUTCOME", "REGRESSION", "count-defaults-retained-for-defense-in-depth", analysis.Summary.DefenseOnlyDefaults, 1, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.unknown-policy-branches", "GUARDRAIL", "COHERENCE", "count-unclassified-schema-value-relations", analysis.Summary.UnknownPolicyBranches, 0, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.repository-writes", "GUARDRAIL", "FOUNDATION", "sum-value-reachability-repository-writes", 0, 0, "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.mutation-authorities", "GUARDRAIL", "FOUNDATION", "join-value-reachability-mutation-authority", 0, 0, "GOVERNOR"),
	}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
