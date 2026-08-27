package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-08-counterexample-first-compiler"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/counterexample-first-compiler.yml",
		"cmd/counterexample-first-compiler-witness",
		"cmd/counterexample-first-judge-witness",
		"docs/research/counterexample-first-meta-compilation.md",
		"examples/counterexample-first-compiler",
		"internal/meta/counterexamplefirst",
		"internal/meta/counterexamplefirstcompiler",
		"internal/meta/counterexamplefirstjudge",
		"internal/verify/scope_counterexample_first_compiler.go",
		"scripts/counterexample-first-compiler",
	}
}
