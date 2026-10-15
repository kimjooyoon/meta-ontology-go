package extractor

import "testing"

func TestFaultReaderDeclarationCapacity(t *testing.T) {
	targets := []string{
		"cmd/gooo#TestGoooFaultOriginReaderUsesSavedFaults",
		"cmd/gooo#assertGoooFaultDescriptions",
		"cmd/gooo#TestGoooFaultOriginReaderDistinguishesFieldUpdates",
		"cmd/gooo#TestFaultOriginPublicFixtureRetainsArchivedRuntime",
		"internal/meta/repositoryprojection/extractor#TestFaultReaderDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
