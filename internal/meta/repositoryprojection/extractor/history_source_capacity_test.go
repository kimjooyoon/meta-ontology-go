package extractor

import "testing"

func TestHistorySourceDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/meta/syntaxregistration#pinPreviousSyntaxTarget",
		"internal/meta/syntaxregistration#readHistoryOrigins",
		"internal/meta/syntaxregistration#historyOwnsHelper",
		"internal/meta/syntaxregistration#historySwitch",
		"internal/meta/syntaxregistration#readHistoryCases",
		"internal/meta/syntaxregistration#historyCaseTarget",
		"internal/meta/syntaxregistration#readHistoryChunks",
		"internal/meta/syntaxregistration#writeHistoryOrigins",
		"internal/meta/syntaxregistration#freshAssertHistoryTarget32403",
		"internal/meta/syntaxregistration#parseHistoryCapacityFixture",
		"internal/meta/syntaxregistration#historyCapacityLegacyFixture",
		"internal/meta/syntaxregistration#TestHistoryCapacityGrowthPreservesEveryDeclaredTarget",
		"internal/meta/syntaxregistration#TestHistoryCapacityRejectsAmbiguousOwnershipAndCycles",
		"internal/meta/syntaxregistration#TestHistoryCapacityCurrentSourceNormalizationProposal",
		"internal/meta/languageassurance/verticalsliceclosureshadow#expectedBoundarySpecsForVersion",
		"internal/meta/languageassurance/verticalsliceclosureshadow#historicalSyntaxTargetChunk0",
		"internal/meta/languageassurance/verticalsliceclosureshadow#historicalSyntaxTargetChunk1",
		"internal/meta/repositoryprojection/extractor#TestHistorySourceDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
