package extractor

import "testing"

func TestSecuritySourceDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/syntax#ReplaceStringLiteral",
		"internal/syntax#replacementSourceSize",
		"internal/syntax#TestLiteralEditPreservesSourceAndEscapesReplacement",
		"internal/syntax#TestLiteralEditRejectsInvalidRangesAndTokens",
		"internal/syntax#TestLiteralReplacementSizeChecksBeforeAllocation",
		"internal/syntax#TestAssemblyIntegerBoundsRetainOnlyValidValues",
		"internal/bodycodegen#replaceActivityProgram",
		"internal/bodycodegen#coerceBodyValue",
		"internal/bodycodegen#coerceBodyInteger",
		"internal/bodycodegen#TestIntegerWidthKeepsTypedWrapSemantics",
		"internal/valueexecution#measure",
		"internal/valueexecution#measurementSources",
		"internal/valueexecution#ProposeSourceRevision",
		"internal/valueexecution#TestMeasurementEditsSelectedLiteralWithRepeatedAndEscapedText",
		"internal/valueexecution#TestMeasurementRejectsStaleAmbiguousAndUnmatchedSources",
		"internal/meta/repositoryprojection/extractor#TestSecuritySourceDeclarationCapacity",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
