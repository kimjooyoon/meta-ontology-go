package verify

const mixedOutcomeTerminalBranch = "agent/mixed-outcome-terminal-v1"

func init() {
	branchScopeAllowlist[mixedOutcomeTerminalBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/language-readiness-witness/rollback-fixed-point",
		"internal/meta/languagereadiness/rollbackfixedpoint",
		"internal/meta/metrictransition",
		"internal/meta/transformationeffect",
		"internal/verify/scope_mixed_outcome_terminal_v1.go",
		"internal/verify/scope_mixed_outcome_terminal_v1_test.go",
	}
}
