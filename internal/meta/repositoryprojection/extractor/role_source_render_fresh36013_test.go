package extractor

import (
	"os"
	"testing"
)

func TestRoleSourceRendering36013(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	required := map[string]bool{
		"freshAssertRoleWriteSet34731": true, "physicalRoleFixture35917": true,
		"TestMultiplePhysicalRoleOrigins34731": true, "assertPhysicalModelOrigins35917": true,
	}
	owners, err := roleDeclarationOwners39247(root, required)
	if err != nil {
		t.Fatal(err)
	}
	checked := map[string]bool{}
	for _, owner := range owners {
		value := observeRenderedCapacity(owner.set, owner.file, owner.raw, owner.function)
		if value.functionStatus != renderedCapacityWithinCap || value.helperStatus != renderedCapacityWithinCap || value.helperLines == nil {
			t.Fatal("role source is not safely renderable", owner.function.Name.Name, value)
		}
		t.Log(owner.function.Name.Name, "source", owner.logical, "rendered", *value.helperLines)
		if !checked[owner.logical] {
			verifyRoleOwnerExtraction39247(t, root, owner.logical)
			checked[owner.logical] = true
		}
	}
	t.Log("all4 verification declarations and actual source partition preserve75-line cap;all9 roles/3 model owners remain;producer/model0")
}
