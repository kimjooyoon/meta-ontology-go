package verify

import "testing"

func TestEntityFieldsV1ScopeIsExplicit(t *testing.T) {
	paths, ok := BranchScope(entityFieldsV1Branch)
	if !ok || len(paths) != 24 { t.Fatalf("EntityFields scope: known=%t paths=%d", ok, len(paths)) }
	allowed := []string{
		".github/workflows/entity-fields-v1.yml", "cmd/entity-fields-witness/main.go", "examples/entity-fields-v1/main.gooo",
		"internal/entityfieldsv1/observation.go", "internal/meta/entityfields/evaluate.go", "internal/syntax/entity_fields_support.go",
	}
	if err := CheckPathScopeForBranch(allowed, entityFieldsV1Branch); err != nil { t.Fatalf("representative paths rejected: %v", err) }
	if err := CheckPathScopeForBranch([]string{".github/workflows/ci.yml"}, entityFieldsV1Branch); err == nil { t.Fatal("unrelated workflow accepted") }
}
