package verify

import (
	"reflect"
	"testing"
)

func TestSyntaxRegistrationOwnershipIsExact(t *testing.T) {
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/repository-projection.yml",
		".github/workflows/syntax-registration-candidate.yml",
		"cmd/syntax-registration-candidate/main.go",
		"cmd/syntax-registration-candidate/output.go",
		"cmd/syntax-registration-candidate/output_test.go",
		"docs/syntax-registration-candidate.md",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/syntaxregistration/candidate_test.go",
		"internal/meta/syntaxregistration/contract.go",
		"internal/meta/syntaxregistration/contract.gooo",
		"internal/meta/syntaxregistration/corpus.go",
		"internal/meta/syntaxregistration/counterexample_test.go",
		"internal/meta/syntaxregistration/denominator.go",
		"internal/meta/syntaxregistration/edits.go",
		"internal/meta/syntaxregistration/execution_identity.go",
		"internal/meta/syntaxregistration/execution_identity_test.go",
		"internal/meta/syntaxregistration/execution_identity_types.go",
		"internal/meta/syntaxregistration/generate.go",
		"internal/meta/syntaxregistration/input.go",
		"internal/meta/syntaxregistration/json.go",
		"internal/meta/syntaxregistration/migration_tests.go",
		"internal/meta/syntaxregistration/native_test.go",
		"internal/meta/syntaxregistration/selection.go",
		"internal/meta/syntaxregistration/source_units.go",
		"internal/meta/syntaxregistration/source_units_test.go",
		"internal/meta/syntaxregistration/syntax_tests.go",
		"internal/meta/syntaxregistration/types.go",
		"internal/verify/scope_syntax_registration_20260907.go",
		"internal/verify/scope_syntax_registration_20260907_test.go",
	}
	got, known := BranchScope("agent/syntax-registration-operator-20260907")
	if !known || !reflect.DeepEqual(got, want) {
		t.Fatalf("registration ownership is not exact: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/syntax-registration-operator-20260907"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".github/workflows/ci.yml", ".github/foundation-authorization.json",
		"internal/meta/generation/registry.go", "internal/meta/semantic/graph.go", "go.mod",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v31.json"} {
		if err := CheckPathScopeForBranch([]string{path}, "agent/syntax-registration-operator-20260907"); err == nil {
			t.Errorf("registration backend ownership allowed unrelated path %q", path)
		}
	}
}
