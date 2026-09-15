package verify

const debuggingUtilityObservedBranch = "agent/debugging-utility-observed-v1"

func init() {
	branchScopeAllowlist[debuggingUtilityObservedBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-utility-evidence.yml",
		"examples/language-debug/contract.json",
		"internal/languagedebug/digest.go",
		"internal/meta/languagedebugexperiment",
		"internal/meta/languageutility",
		"scripts/language-debug-experiment",
		"scripts/language-utility-evidence",
		"internal/verify/scope_debugging_utility_observed_v1.go",
		"internal/verify/scope_debugging_utility_observed_v1_test.go",
	}
}
