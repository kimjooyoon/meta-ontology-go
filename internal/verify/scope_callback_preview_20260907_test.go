package verify

import (
	"reflect"
	"testing"
)

func TestCallbackPreviewScopeHasExactRepairSurface(t *testing.T) {
	want := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"cmd/callback-preview/factory_test.go",
		"cmd/callback-preview/main.go",
		"cmd/callback-preview/main_test.go",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/generation/callback-extraction-contract.gooo",
		"internal/meta/generation/callback-preview-contract.gooo",
		"internal/meta/generation/callback_extraction_contract.go",
		"internal/meta/generation/callback_extraction_contract_test.go",
		"internal/meta/generation/callback_extraction_observer_test.go",
		"internal/meta/generation/callback_preview_contract.go",
		"internal/meta/generation/callback_preview_contract_test.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/repositoryprojection/extractor/callback_extraction_claims.go",
		"internal/meta/repositoryprojection/extractor/callback_extraction_observation.go",
		"internal/meta/repositoryprojection/extractor/callback_extraction_observation_test.go",
		"internal/meta/repositoryprojection/extractor/callback_extraction_plan.go",
		"internal/meta/repositoryprojection/extractor/callback_extraction_plan_test.go",
		"internal/meta/repositoryprojection/extractor/callback_factory.go",
		"internal/meta/repositoryprojection/extractor/callback_factory_test.go",
		"internal/meta/repositoryprojection/extractor/callback_observation_module.go",
		"internal/meta/repositoryprojection/extractor/callback_observation_module_test.go",
		"internal/meta/repositoryprojection/extractor/callback_observation_workspace.go",
		"internal/meta/repositoryprojection/extractor/callback_preview.go",
		"internal/meta/repositoryprojection/extractor/callback_preview_test.go",
		"internal/meta/repositoryprojection/extractor/capacity_failure_diagnostics_test.go",
		"internal/meta/repositoryprojection/extractor/closure_preservation.go",
		"internal/meta/repositoryprojection/extractor/closure_preservation_structure.go",
		"internal/meta/repositoryprojection/extractor/closure_preservation_test.go",
		"internal/meta/repositoryprojection/extractor/decompose.go",
		"internal/meta/repositoryprojection/extractor/extract.go",
		"internal/meta/repositoryprojection/extractor/rendered_capacity_progress_regression_test.go",
		"internal/meta/repositoryprojection/extractor/suffix_closure_identity_test.go",
		"internal/verify/scope_callback_preview_20260907.go",
		"internal/verify/scope_callback_preview_20260907_test.go",
	}
	got, known := BranchScope("agent/main-language-preview-20260906")
	if !known || len(got) != 38 || !reflect.DeepEqual(got, want) {
		t.Fatalf("callback preview ownership is not exact: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/main-language-preview-20260906"); err != nil {
		t.Fatal(err)
	}
}

func TestCallbackPreviewScopeDoesNotGrantUnrelatedAuthority(t *testing.T) {
	denied := []string{
		".github/foundation-authorization.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/gooo-release-publish.yml",
		".github/ci-governance.json.backup",
		"go.mod",
		"cmd/callback-observe/main.go",
		"cmd/gooo/main_part01.go",
		"internal/valueexecution/record_input.go",
		"internal/meta/repositoryprojection/extractor/callback_preview_extra.go",
		"cmd/language-readiness-witness/predecessor-selection/pagination_test.go",
	}
	for _, path := range denied {
		if err := CheckPathScopeForBranch([]string{path}, "agent/main-language-preview-20260906"); err == nil {
			t.Errorf("callback preview scope allowed unrelated path %q", path)
		}
	}
	for _, branch := range []string{"agent/main-language-preview-20260906-other", "agent/unknown"} {
		if err := CheckPathScopeForBranch([]string{"cmd/callback-preview/main.go"}, branch); err == nil {
			t.Errorf("unregistered branch %q inherited callback preview scope", branch)
		}
	}
}
