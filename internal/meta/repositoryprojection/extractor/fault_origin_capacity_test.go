package extractor

import "testing"

func TestFaultOriginDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/bodyexecution#RecordedFaultLocalRHS",
		"internal/bodyexecution#recordedFaultFact",
		"internal/bodyexecution#nativeOriginPage.visitFaults",
		"internal/bodyexecution#validateRecordedFaultOrigins",
		"internal/bodyexecution#ValidateNativeOriginQuery",
		"internal/bodyexecution#selectRecordedNativeOrigins",
		"internal/bodyexecution#nativeOriginPage.visit",
		"internal/bodyexecution#validateOriginSubjectSelection",
		"cmd/gooo#packageFaultLocalRHS",
		"cmd/gooo#savedFaultOriginRuntime",
		"cmd/gooo#TestFaultOriginSelectionPreservesRecordedOperations",
		"cmd/gooo#TestFaultOriginSelectionRetainsMixedAndEmptyTables",
		"cmd/gooo#TestFaultOriginRecordedContractRejectsBrokenLinks",
		"cmd/gooo#TestFaultOriginReadIsNotAProducerReexecution",
		"internal/meta/repositoryprojection/extractor#TestFaultOriginDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
