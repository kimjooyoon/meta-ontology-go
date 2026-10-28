package extractor

import "testing"

func TestRecordIterationSourceRenderedCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodyexecution#compositionFieldProjection.bindSources",
		"internal/bodyexecution#compositionFieldProjection.recordSources",
		"internal/bodyexecution#compositionFieldProjection.bindRangeFieldSources",
		"internal/bodyexecution#validateFieldSources",
		"internal/bodyexecution#fieldSourceBodySHA",
		"internal/bodyexecution#validateFieldSourceOrigin",
		"internal/bodyexecution#validateDerivedRangeFieldSource",
		"internal/bodyexecution#validateSelectedFieldSource",
	})
}
