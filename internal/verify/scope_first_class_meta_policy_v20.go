package verify

const firstClassMetaPolicyV20Branch = "agent/first-class-meta-policy-v20"

func init() {
	branchScopeAllowlist[firstClassMetaPolicyV20Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		"cmd/meta-policy-compilation-consumer/**",
		"docs/language/meta-policy-compilation.md",
		"examples/meta-policy-compilation/**",
		"internal/bidir/**",
		"internal/meta/policycompilation/**",
		"internal/semantic/**",
		"internal/syntax/**",
		"internal/verify/scope_first_class_meta_policy_v20.go",
	}
}
