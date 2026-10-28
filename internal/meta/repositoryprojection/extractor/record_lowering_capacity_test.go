package extractor

import (
	"os"
	"testing"
)

func TestRecordDocumentLoweringRenderedCapacity(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	requireNamedCapacities(t, root, []string{
		"internal/bidir#lowerDocumentContextWithTypesAndEntityFieldsSupport",
		"internal/bidir#lowerDocumentPolicies",
	})
}
