package extractor

import "testing"

func TestRecordCompoundDeclarationCapacity(t *testing.T) {
	targets := []string{
		"internal/bodycodegen#bodyProjection.captureFieldOrigins",
		"internal/bodycodegen#WalkFieldSyntax",
		"internal/bodycodegen#validateBodyStatement",
		"internal/bodycodegen#integerBodyEvaluator.evaluateAssignment",
		"internal/bodycodegen#integerBodyEvaluator.assignRecordField",
		"internal/bodyexecution#arithmeticProjection.collect",
		"internal/bodyexecution#arithmeticProjection.render",
		"internal/bodyexecution#validateFieldBinding",
		"internal/bodyexecution#validateFieldInventory",
		"internal/bodyexecution#compositionFieldProjection.collectFunctionWrites",
		"internal/bodyexecution#compositionFieldProjection.collectFunctionReads",
		"internal/bodyexecution#compositionFieldInventory.addParameters",
		"internal/bodyexecution#validateNativeFieldCase",
		"internal/bodyexecution#observedCompositionFieldProjection",
		"internal/bodyexecution#observedCompositionFieldProjectionWithFaults",
		"internal/bodyexecution#compositionFieldRenderer.node",
		"internal/bodyexecution#compositionFieldRenderer.scalarWrite",
		"internal/bodyexecution#compositionFieldRenderer.scalarValue",
		"internal/bodyexecution#compositionFieldProjection.bindSources",
		"internal/bodyexecution#observedCompositionLocalEvidence",
		"internal/bodyexecution#fieldFact",
		"internal/bodyexecution#fillFieldFactWrite",
		"internal/bodyexecution#validateFieldSources",
		"internal/packageruntime/workspaceexecution#isActivityInput",
		"cmd/gooo#TestGoooReadsCompoundFieldOrigins",
		"cmd/gooo#readCompoundRuntimeFixture",
		"cmd/gooo#TestGoooReadsCompoundTitleOrigin",
		"internal/bodycodegen#assertRecordCompoundFlow",
		"internal/bodycodegen#TestRecordCompoundRefusesInvalidTargets",
		"internal/bodycodegen#TestRecordCompoundGenerationAndEvaluation",
		"internal/bodycodegen#TestRecordCompoundSourceOrigins",
		"internal/bodyexecution#singleAssignmentTarget",
		"internal/bodyexecution#compositionFieldProjection.collectFieldAssignment",
		"internal/bodyexecution#compositionFieldRenderer.compoundWrite",
		"internal/bodyexecution#arithmeticProjection.collectFieldDivision",
		"internal/bodyexecution#compositionFieldProjection.bindFieldFaults",
		"internal/bodyexecution#validateFieldOperation",
		"internal/bodyexecution#validateCompoundFieldReadWindow",
		"internal/bodyexecution#TestNamedRecordPortsRetainPositions",
		"internal/bodyexecution#TestRecordCompoundDivisionFaultOrigin",
		"internal/bodyexecution#assertInterruptedCompoundWrite",
		"internal/bodyexecution#assertRecordCompoundJournal",
		"internal/bodyexecution#TestCompoundFieldWindowRequiresPreviousValue",
		"internal/bodyexecution#TestRecordCompoundNativeOrigins",
		"internal/bodyexecution#recordCompoundNativeCases",
		"internal/bodyexecution#TestRecordCompoundNestedFaults",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range targets {
		r := observed[target].capacity
		if r.functionStatus != renderedCapacityWithinCap || r.helperStatus != renderedCapacityWithinCap || r.helperLines == nil {
			t.Errorf("%s: %+v", target, r)
		} else {
			t.Log(target, "rendered", *r.helperLines)
		}
	}
}
