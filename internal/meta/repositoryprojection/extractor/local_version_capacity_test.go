package extractor

import "testing"

// The origin adapter is projected through the same bounded helper representation.
func TestLocalVersionAdapterDeclarationCapacity(t *testing.T) {
	targets := []string{
		"cmd/gooo#packageOriginQuery",
		"cmd/gooo#configurePackageOriginShape",
		"cmd/gooo#validatePackageOriginShape",
		"cmd/gooo#packageOriginInputsWithShape",
		"cmd/gooo#encodePackageOriginInputs",
		"internal/bodyexecution#SelectRecordedLocalVersions",
		"internal/bodyexecution#newLocalVersionColumns",
		"internal/bodyexecution#LocalVersionColumns.appendWrite",
	}
	requireNamedCapacities(t, "../../../..", targets)
}
