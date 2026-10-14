package extractor

import (
	"os"
	"testing"
)

func TestCounterSurfaceCapacity19823(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{
		"internal/bodycodegen#rewriteBodySurface",
		"internal/bodycodegen#bodySurfaceMap.originalOffset",
		"internal/bodycodegen#bodySurfaceMap.originalRange",
		"internal/bodycodegen#NormalizeBodySource",
		"internal/bodycodegen#BodySourceMap.OriginalRange",
		"internal/bodycodegen#scanBodySurface",
		"internal/bodycodegen#BodySourceMap.NormalizedRange",
		"internal/bodycodegen#bodySurfaceMap.normalizedOffset",
		"internal/bodycodegen#countedLoopShape",
		"internal/bodycodegen#loopCounter",
		"internal/bodycodegen#loopIntegerLiteral",
		"internal/bodycodegen#validateLoopCounterBody",
		"internal/bodycodegen#validateCountedLoop",
		"internal/bodycodegen#loopUnitForwardStep",
		"internal/bodycodegen#integerBodyEvaluator.evaluateCountedLoop",
		"internal/bodycodegen#integerBodyEvaluator.initializeCountedLoop",
		"internal/bodycodegen#integerBodyEvaluator.evaluateCountedPost",
		"internal/bodycodegen#normalizeCountedInitializers",
		"internal/bodyexecution#readProjection.localSlot",
		"internal/bodyexecution#readProjection.localAssignment",
		"internal/bodyexecution#readProjection.renderLocal",
		"internal/packageruntime/workspaceexecution#localFileLocation",
		"internal/packageruntime/workspaceexecution#retainedLocalTemplateRange",
		"internal/packageruntime/workspaceexecution#localFileConstruction",
		"internal/packageruntime/workspaceexecution#fileExpressionConstruction",
	}
	requireNamedCapacities(t, root, targets)
	t.Log("required declaration population", len(targets), "unchanged75-line generated helper cap;package/receiver/function identity survives logical source splits;producer/model0")
}
