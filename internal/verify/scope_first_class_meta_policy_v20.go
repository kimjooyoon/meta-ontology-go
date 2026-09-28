package verify

const firstClassMetaPolicyV20Branch = "agent/first-class-meta-policy-v20"

func init() {
	branchScopeAllowlist[firstClassMetaPolicyV20Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		"bootstrap/repository-projector/evidence",
		"cmd/meta-policy-compilation-consumer",
		"docs/language/meta-policy-compilation.md",
		"examples/meta-policy-compilation",
		"examples/repository-projection-diagnostics/package-partition.json",
		"internal/bidir",
		"internal/meta/languagereadiness/languagesyntax/replay/shape.go",
		"internal/meta/policycompilation",
		"internal/semantic",
		"internal/syntax",
		"internal/verify/scope_first_class_meta_policy_v20.go",
	}
}
