package verify

import "testing"

func TestDebuggingUtilityObservedScope(t *testing.T) {
	paths, ok := BranchScope(debuggingUtilityObservedBranch)
	if !ok || len(paths) != 11 {
		t.Fatalf("debugging utility branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/language-utility-evidence.yml",
		"examples/language-debug/contract.json",
		"internal/meta/languagedebugexperiment/evaluate.go",
		"internal/meta/languageutility/observation.go",
		"scripts/language-debug-experiment/main.sh",
		"scripts/language-utility-evidence/main.sh",
		"internal/verify/scope_debugging_utility_observed_v1.go",
		"internal/verify/scope_debugging_utility_observed_v1_test.go",
		".github/agent-scope-table.md",
		".github/ci-governance.json",
	}
	if err := CheckPathScopeForBranch(allowed, debuggingUtilityObservedBranch); err != nil {
		t.Fatalf("representative owned paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"docs/unrelated.md"}, debuggingUtilityObservedBranch); err == nil {
		t.Fatal("unrelated path was accepted")
	}
}
