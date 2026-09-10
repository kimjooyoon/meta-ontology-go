package verify

func init() {
	branchScopeAllowlist["agent/syntax-negative-evidence-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"internal/verify/scope_syntax_negative_evidence_20260911.go",
		"cmd/language-syntax-witness/run.go",
		"cmd/language-syntax-witness/io.go",
		"cmd/language-syntax-witness/negative_report.go",
		"cmd/language-syntax-witness/negative_report_test.go",
	}
}
