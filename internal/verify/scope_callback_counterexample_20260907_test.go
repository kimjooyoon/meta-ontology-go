package verify

import (
	"reflect"
	"testing"
)

func TestCallbackCounterexampleScopeIsExact(t *testing.T) {
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/repositoryprojection/extractor/callback_extraction_observation.go",
		"internal/meta/repositoryprojection/extractor/callback_counterexample_retention_test.go",
		"internal/meta/repositoryprojection/extractor/callback_counterexample_claim_test.go",
		"internal/verify/scope_callback_counterexample_20260907.go",
		"internal/verify/scope_callback_counterexample_20260907_test.go",
	}
	got, known := BranchScope("agent/callback-counterexample-retention-20260907")
	if !known || !reflect.DeepEqual(got, want) {
		t.Fatalf("callback counterexample scope changed: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/callback-counterexample-retention-20260907"); err != nil {
		t.Fatal(err)
	}
	for _, denied := range []string{".github/workflows/ci.yml", ".github/foundation-authorization.json",
		"internal/meta/generation/callback-extraction-contract.gooo",
		"cmd/language-readiness-witness/predecessor-selection/pagination_test.go", "go.mod"} {
		if err := CheckPathScopeForBranch([]string{denied}, "agent/callback-counterexample-retention-20260907"); err == nil {
			t.Errorf("callback counterexample allowed unrelated path %q", denied)
		}
	}
}
