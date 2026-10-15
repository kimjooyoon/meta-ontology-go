package extractor

import "testing"

func TestFaultDiagnosticDeclarationCapacity(t *testing.T) {
	targets := []string{
		"cmd/gooo#writePackageEntryValues",
		"cmd/gooo#writePackageFaultContext",
		"cmd/gooo#packageFaultLocalRHS",
		"cmd/gooo#savedPackageDivisionDiagnostic",
		"cmd/gooo#TestPackageFaultContextUsesSavedGoooOrigins",
		"cmd/gooo#TestPackageFaultContextKeepsUnmappedOriginsExplicit",
		"internal/meta/repositoryprojection/extractor#TestFaultDiagnosticDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
