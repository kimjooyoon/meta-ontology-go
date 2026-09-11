package verify

const metaPolicyCompilationSemanticAuthorityBranch = "agent/meta-policy-compilation-semantic-authority-20260901"

func init() {
	branchScopeAllowlist[metaPolicyCompilationSemanticAuthorityBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		".github/workflows/transformation-effect.yml",
		"cmd/meta-policy-compilation-consumer",
		"cmd/meta-policy-compilation-witness",
		"docs/language/meta-policy-compilation.md",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/meta-policy-compilation",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/policycompilation",
		"internal/verify/scope_meta_policy_compilation_semantic_authority_20260901.go",
	}
}

const metaPolicyPublicProfileBranch = "agent/meta-policy-public-profile-20260911"

func init() {
	branchScopeAllowlist[metaPolicyPublicProfileBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-policy-compilation.yml",
		"cmd/gooo/generate_part01.go",
		"cmd/gooo/generate_pipeline_part03.go",
		"cmd/gooo/meta_policy_profile.go",
		"cmd/meta-policy-compilation-consumer/main.go",
		"cmd/meta-policy-compilation-witness/main.go",
		"docs/language/meta-policy-compilation.md",
		"internal/meta/policycompilation/compile.go",
		"internal/meta/policycompilation/evaluate.go",
		"internal/meta/policycompilation/judge.go",
		"internal/meta/policycompilation/model.go",
		"internal/meta/policycompilation/profile.go",
		"internal/meta/policycompilation/receipt.go",
		"internal/meta/policycompilation/strict_json.go",
		"internal/meta/policycompilation/strict_json_test.go",
		"internal/meta/policycompilation/verify.go",
		"internal/verify/scope_meta_policy_compilation_semantic_authority_20260901.go",
	}
}
