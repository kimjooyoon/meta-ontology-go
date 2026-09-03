package verify

const compilerSelfImprovement20260903Branch = "agent/compiler-self-improvement-20260903"
const compilerEnvelopeSelfImprovement20260903Branch = "agent/compiler-envelope-self-improvement-20260903"

func init() {
	branchScopeAllowlist[compilerSelfImprovement20260903Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		".github/workflows/transformation-effect.yml",
		"examples/compiler-self-improvement",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/generator/normalize_part01.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v24.json",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
	}
	branchScopeAllowlist[compilerEnvelopeSelfImprovement20260903Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		"examples/compiler-self-improvement",
		"internal/generator/types_part01.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
	}
}
