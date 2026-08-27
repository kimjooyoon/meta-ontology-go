package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-18-policy-compilation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		"cmd/meta-policy-compilation-witness",
		"cmd/meta-policy-compilation-consumer",
		"docs/language/meta-policy-compilation.md",
		"examples/meta-policy-compilation",
		"internal/meta/policycompilation",
		"internal/verify/scope_policy_compilation.go",
	}
}
