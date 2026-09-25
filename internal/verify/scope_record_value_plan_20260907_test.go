package verify

import (
	"reflect"
	"testing"
)

func TestRecordValuePlanScopeHasExactRepairSurface(t *testing.T) {
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"cmd/gooo/run_source_part01.go",
		"cmd/gooo/run_source_part02.go",
		"cmd/gooo/run_source_record.go",
		"cmd/gooo/run_source_record_test.go",
		"examples/language-record-binding/README.md",
		"examples/language-record-binding/input.json",
		"examples/language-record-binding/main.gooo",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator_migration_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v30.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/valueexecution/record_bindings.go",
		"internal/valueexecution/record_compile.go",
		"internal/valueexecution/record_execute.go",
		"internal/valueexecution/record_input.go",
		"internal/valueexecution/record_plan_test.go",
		"internal/valueexecution/record_result.go",
		"internal/verify/scope_record_value_plan_20260907.go",
		"internal/verify/scope_record_value_plan_20260907_test.go",
	}
	got, known := BranchScope("agent/record-value-plan-20260907")
	if !known || len(got) != 27 || !reflect.DeepEqual(got, want) {
		t.Fatalf("record plan ownership is not exact: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/record-value-plan-20260907"); err != nil {
		t.Fatal(err)
	}
}

func TestRecordValuePlanScopeDoesNotGrantUnrelatedAuthority(t *testing.T) {
	denied := []string{
		".github/foundation-authorization.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/gooo-release-publish.yml",
		".github/ci-governance.json.backup",
		"go.mod",
		"internal/semantic/graph.go",
		"internal/valueexecution/record_input_extra.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v31.json",
	}
	for _, path := range denied {
		if err := CheckPathScopeForBranch([]string{path}, "agent/record-value-plan-20260907"); err == nil {
			t.Errorf("record plan scope allowed unrelated path %q", path)
		}
	}
	for _, branch := range []string{"agent/record-value-plan-20260907-other", "agent/unknown"} {
		if err := CheckPathScopeForBranch([]string{"internal/valueexecution/record_input.go"}, branch); err == nil {
			t.Errorf("unregistered branch %q inherited record plan scope", branch)
		}
	}
}
