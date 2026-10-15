package extractor

import "testing"

func TestRecordedUsageDeclarationCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"cmd/gooo#packageOriginInputsWithShape", "cmd/gooo#configurePackageOriginShape", "cmd/gooo#validatePackageOriginShape",
		"internal/bodyexecution#SelectRecordedLocalUses", "internal/bodyexecution#localUseTraceIndices",
		"internal/bodyexecution#newLocalUseColumns", "internal/bodyexecution#LocalUseColumns.appendUse",
	})
}
