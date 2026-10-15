package extractor

import "testing"

func TestLocalDivisionDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/bodyexecution#arithmeticCompositionProjection",
		"internal/bodyexecution#arithmeticProjection.collect",
		"internal/bodyexecution#arithmeticProjection.collectLocalDivision",
		"internal/bodyexecution#compositionFieldProjection.bindFieldFaults",
		"internal/bodyexecution#observedCompositionLocalEvidence",
		"internal/bodyexecution#observedReadProjection",
		"internal/bodyexecution#observedReadProjectionWithLocals",
		"internal/bodyexecution#readProjection.bindLocalFaults",
		"internal/bodyexecution#readProjection.renderLocal",
		"internal/bodyexecution#TestLocalDivisionPreservesFaultOrigins",
		"internal/bodyexecution#localDivisionFaultCases",
		"internal/bodyexecution#assertLocalDivisionInterrupted",
		"internal/bodyexecution#TestMixedLocalFieldDivisionsRetainSeparateOrigins",
		"internal/bodyexecution#TestAuthoredLocalDivisionExample",
		"internal/bodyexecution#TestLocalDivisionFaultBindingRequiresExactVariable",
		"internal/meta/repositoryprojection/extractor#TestLocalDivisionDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}

func TestLocalDivisionObservationCapacity(t *testing.T) {
	targets := []string{
		"cmd/gooo#TestLocalDivisionPublicRegistryProfile",
		"cmd/gooo#TestLocalDivisionSavedProvenance",
		"internal/meta/repositoryprojection/extractor#TestLocalDivisionObservationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
