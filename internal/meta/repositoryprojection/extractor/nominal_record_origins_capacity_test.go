package extractor

import "testing"

func TestNominalRecordOriginsRenderedCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodycodegen#recordValueFlow",
		"internal/bodycodegen#recordFlowBuilder.parse",
		"internal/bodycodegen#recordFlowBuilder.wrap",
		"internal/bodycodegen#recordFlowBuilder.selector",
		"internal/bodycodegen#recordFlowBuilder.constructor",
		"internal/bodycodegen#recordFlowBuilder.binary",
		"internal/bodycodegen#recordFlowBuilder.assignment",
		"internal/bodycodegen#recordFlowBuilder.joinStates",
		"internal/bodycodegen#recordFlowBuilder.call",
		"internal/bodycodegen#recordFlowBuilder.markCallee",
		"internal/bodycodegen#recordFlowBuilder.retainReturn",
		"internal/bodycodegen#InspectTypedValueOrigins",
		"internal/bodycodegen#recordFlowBuilder.typedOriginOutputs",
		"internal/bodycodegen#recordFlowBuilder.finishTypedOrigins",
		"internal/bodycodegen#typedOriginSources",
		"internal/bodycodegen#TypedValueOriginInputs",
		"internal/bodycodegen#TypedSourceControlInputs",
		"internal/bodycodegen#writeOriginTextFacts",
		"internal/bodycodegen#recordFlowBuilder.nominalValue",
		"internal/bodycodegen#recordFlowBuilder.retainNominalValue",
		"internal/bodycodegen#recordFlowBuilder.inputRecordFields",
		"internal/bodycodegen#recordFlowBuilder.recordFieldRoot",
		"internal/bodycodegen#recordFlowBuilder.nominalRoot",
		"internal/bodycodegen#recordFlowBuilder.joinField",
		"internal/bodycodegen#recordFlowBuilder.selectorBase",
		"internal/bodycodegen#recordFlowBuilder.selectorValue",
		"internal/bodycodegen#recordFlowUsesNominalValues",
		"internal/bodycodegen#recordFlowBuilder.linkRecordVersion",
		"internal/bodycodegen#recordFlowBuilder.exportRecordVersions",
		"internal/bodycodegen#recordFlowNeedsVersionFeatures",
		"internal/bodycodegen#recordOriginContext",
		"internal/bodycodegen#recordGraphContext",
	})
}
