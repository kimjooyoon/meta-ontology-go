package verify

import "testing"

func TestMixedOutcomeTerminalScope(t *testing.T) {
	paths, ok := BranchScope(mixedOutcomeTerminalBranch)
	if !ok || len(paths) != 8 {
		t.Fatalf("mixed outcome branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		"cmd/language-readiness-witness/rollback-fixed-point/run.go",
		"internal/meta/languagereadiness/rollbackfixedpoint/build.go",
		"internal/meta/metrictransition/outcome.go",
		"internal/meta/transformationeffect/document_types.go",
	}
	if err := CheckPathScopeForBranch(allowed, mixedOutcomeTerminalBranch); err != nil {
		t.Fatalf("representative mixed outcome paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"internal/meta/metrictransition-other"}, mixedOutcomeTerminalBranch); err == nil {
		t.Fatal("unregistered metric-transition path was accepted")
	}
}
