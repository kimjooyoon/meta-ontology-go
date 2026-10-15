package extractor

import "testing"

func TestFaultArchiveDeclarationCapacity(t *testing.T) {
	targets := []string{
		"cmd/gooo#TestFaultOriginSavedSourceAndMembership",
		"internal/meta/repositoryprojection/extractor#TestFaultArchiveDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
