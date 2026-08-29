package verify

import "testing"

func TestBaseRepairRepositoryProjectionScope(t *testing.T) {
	paths, ok := BranchScope(baseRepairRepositoryProjectionBranch)
	if !ok || len(paths) != 30 {
		t.Fatalf("base repair branch was not registered exactly: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/ci.yml",
		"bootstrap/function-extractor/apply.go",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/sourcepolicy/indicator.go",
		"scripts/meta-summary/render.go",
		"internal/verify/scope_base_repair_repository_projection.go",
	}
	if err := CheckPathScopeForBranch(allowed, baseRepairRepositoryProjectionBranch); err != nil {
		t.Fatalf("representative owned paths were rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"docs/unrelated.md"}, baseRepairRepositoryProjectionBranch); err == nil {
		t.Fatal("unrelated path was accepted")
	}
}
