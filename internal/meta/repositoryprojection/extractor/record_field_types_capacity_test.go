package extractor

import (
	"os"
	"testing"
)

func TestSharedRecordFieldTypesRenderedCapacity(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	requireNamedCapacities(t, root, []string{
		"internal/semantic#ValidateRecordReferences",
		"internal/semantic#recordReferenceGraph.visit",
		"internal/bidir#validateRecordReferenceGraph",
		"internal/bidir#Model.ValidateWithEntityFieldsSupport",
		"internal/generator#EntityFieldTypeNames",
		"internal/generator#projectedEntityFieldType",
		"internal/generator#validateRecordFieldReferences",
		"internal/generator#validateSupportedEntityFields",
		"internal/generator#validateSupportedField",
		"internal/generator#prepareEntityFields",
		"internal/generator#Generator.generateWithEntityFieldsSupport",
		"internal/generator#generateProjectionV1WithEntityFieldsSupport",
		"cmd/gooo#validateCLIProjectedFields",
		"cmd/gooo#projectionIRFromBidirModelWithSupport",
		"cmd/gooo#projectionBidirFieldTypeID",
	})
}
