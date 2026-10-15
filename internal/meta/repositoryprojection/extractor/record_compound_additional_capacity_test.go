package extractor

import "testing"

func TestRecordCompoundAdditionalDeclarationCapacity(t *testing.T) {
	targets := []string{
		"cmd/gooo#evaluateRoundTrip",
		"cmd/gooo#runWithInputCommandsTwo",
		"cmd/gooo#TestRecordCompoundExampleSourceCoherence",
		"cmd/gooo#TestRecordCompoundExampleWorkspaceResolution",
		"cmd/gooo#TestRecordCompoundSavedOrigins",
		"internal/bodyexecution#TestCompoundFieldFaultBindingRequiresCompleteOriginSet",
		"internal/syntax#TestStage0SyntaxContractDiagnosticsAreDeterministic",
		"internal/syntax#TestParseDiagnosticsTable",
		"internal/meta/repositoryprojection/extractor#TestRecordCompoundDeclarationCapacity",
		"internal/meta/repositoryprojection/extractor#TestRecordCompoundAdditionalDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}

func TestRecordExampleRelocationCapacity(t *testing.T) {
	observed, err := collectNamedCapacityFiles("../../../..", []string{
		"cmd/gooo#TestRegisteredRecordExamplePreservesProjection",
		"cmd/gooo#TestRecordCompoundSavedOrigins",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, []string{
		"cmd/gooo#TestRegisteredRecordExamplePreservesProjection",
		"cmd/gooo#TestRecordCompoundSavedOrigins",
	}, observed)
}
