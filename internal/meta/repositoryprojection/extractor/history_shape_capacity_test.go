package extractor

import "testing"

func TestHistoryShapeDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/meta/syntaxregistration#isHistorySelector",
		"internal/meta/syntaxregistration#isHistoryCall",
		"internal/meta/syntaxregistration#historyHelperSignature",
		"internal/meta/syntaxregistration#historyContinuation",
		"internal/meta/syntaxregistration#readHistoryChunks",
		"internal/meta/syntaxregistration#TestHistorySelectorUsesSyntaxInsteadOfLineLayout",
		"internal/meta/syntaxregistration#TestPartitionedHistoryContinuesNineRegistrationRoles",
		"internal/meta/syntaxregistration#historyContinuationSource",
		"internal/meta/repositoryprojection/extractor#TestHistoryShapeDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
