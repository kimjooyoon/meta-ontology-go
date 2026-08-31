package verify

const metaPolicyCompilationSemanticAuthorityBranch = "agent/meta-policy-compilation-semantic-authority-20260901"

func init() {
	branchScopeAllowlist[metaPolicyCompilationSemanticAuthorityBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		"cmd/meta-policy-compilation-consumer",
		"cmd/meta-policy-compilation-witness",
		"docs/language/meta-policy-compilation.md",
		"examples/meta-policy-compilation",
		"internal/meta/policycompilation",
		"internal/verify/scope_meta_policy_compilation_semantic_authority_20260901.go",
	}
}
