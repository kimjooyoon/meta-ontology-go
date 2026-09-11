package verify

const firstClassMetaPolicyV201Branch = "agent/first-class-meta-policy-v20-1"

func init() {
	branchScopeAllowlist[firstClassMetaPolicyV201Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/verify/scope_first_class_meta_policy_v20_1.go",
	}
}
