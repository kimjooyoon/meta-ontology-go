package verify

import (
	"reflect"
	"testing"
)

const goooSingleActivityCompositionBranch = "agent/gooo-single-activity-composition-v1-20261007"

func TestGoooSingleActivityCompositionScopeIsExact(t *testing.T) {
	paths, ok := BranchScope(goooSingleActivityCompositionBranch)
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/body_compose_test.go",
		"docs/language-direction.ko.md",
		"docs/native-body-composition.md",
		"examples/body-codegen/native-single-activity-cases.json",
		"examples/body-codegen/native-single-activity.gooo.fixture",
		"internal/bidir/typed_plan.go",
		"internal/bidir/typed_plan_test.go",
		"internal/bodyexecution/composition_prepare.go",
		"internal/verify/scope_gooo_single_activity_composition_20261007.go",
		"internal/verify/scope_gooo_single_activity_composition_20261007_test.go",
	}
	if !ok || !reflect.DeepEqual(paths, want) {
		t.Fatalf("scope = %v, known=%t; want exact paths %v", paths, ok, want)
	}
}
