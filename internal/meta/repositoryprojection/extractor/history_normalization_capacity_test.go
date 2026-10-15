package extractor

import "testing"

func TestHistoryNormalizationOwnershipDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/meta/syntaxregistration#historyNormalizationUnits",
		"internal/meta/syntaxregistration#historyNormalizationLayout",
		"internal/meta/syntaxregistration#TestHistoryNormalizationFollowsDeclarationOwners",
		"internal/meta/syntaxregistration#TestHistoryCapacityCurrentSourceNormalizationProposal",
		"internal/meta/repositoryprojection/extractor#TestHistoryNormalizationOwnershipDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
