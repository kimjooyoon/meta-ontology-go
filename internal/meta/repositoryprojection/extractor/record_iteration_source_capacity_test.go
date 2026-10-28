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
		"internal/bodyexecution#bindBridgeElementReads",
		"internal/bodyexecution#iterationFieldSources",
		"internal/bodyexecution#validateBridgeElementReads",
		"internal/bodyexecution#sameIterationBinding",
		"internal/bodyexecution#validateBridgeElementSource",
		"internal/bodyexecution#hasBridgeElementReads",
		"internal/bodyexecution#fillFieldElementFact",
		"internal/bodyexecution#compositionFieldRenderer.rangeRecordWrite",
		"internal/bodyexecution#pairedRangeRead",
		"internal/bodyexecution#bridgeIterationHooks",
		"internal/bodyexecution#summarizeBridgeElementReads",
		"internal/bodyexecution#bindBridgeElementSummary",
		"internal/bodyexecution#validateBridgeObservations",
		"internal/bodyexecution#validateBridgeTrace",
		"internal/bodyexecution#bindBridgeFile",
	})
}
