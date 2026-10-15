package extractor

import "testing"

func TestNamedReaderDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/bodycodegen#TypedNamedInputDeclarationInputs",
		"internal/bodycodegen#inputColumnsFromRows",
		"internal/bodycodegen#marshalDeclarationInputs",
		"internal/bodycodegen#TestNamedColumnsSourceTransport",
		"internal/bodycodegen#TestNamedColumnsRejectSourceDisagreement",
		"cmd/gooo#TestNamedSignatureExportCLI",
		"cmd/gooo#TestNamedInputReaderFromCompilerFacts",
		"cmd/gooo#namedReaderCommand",
		"cmd/gooo#assertNamedReaderResults",
		"cmd/gooo#TestNamedReaderSavedSourceBinding",
		"cmd/gooo#readNamedReaderObservation",
		"internal/meta/repositoryprojection/extractor#collectNamedCapacityFiles",
		"internal/meta/repositoryprojection/extractor#requireObservedCapacities",
		"internal/meta/repositoryprojection/extractor#TestNamedCapacityIdentityAfterFileMove",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
