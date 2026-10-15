package extractor

import (
	"os"
	"testing"
)

func TestLengthIntegerCapacity20311(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{
		"internal/bodycodegen#normalizeLengthIntegerContexts",
		"internal/bodycodegen#lengthIntegerContext.widen",
		"internal/bodycodegen#isImplicitLengthConversion",
		"internal/bodycodegen#runtimeLengthBindings",
		"internal/bodycodegen#lengthIntegerContext.visit",
		"internal/bodycodegen#lengthIntegerContext.arguments",
		"internal/bodycodegen#normalizeIntegerLocalInitializers",
		"internal/bodycodegen#recordFlowBuilder.expression",
	}
	requireNamedCapacities(t, root, targets)
	t.Log("required declaration population", len(targets), "unchanged75-line generated helper cap;package/receiver/function identity survives logical source splits;producer/model0")
}
