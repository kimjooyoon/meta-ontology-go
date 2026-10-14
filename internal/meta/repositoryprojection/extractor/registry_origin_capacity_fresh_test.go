package extractor

import (
	"os"
	"testing"
)

func TestRegistryOriginCapacity29111(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{
		"internal/meta/syntaxregistration#generateRegistry",
		"internal/meta/syntaxregistration#goSource.registryCaseTail",
		"internal/meta/syntaxregistration#namedRegistryType",
		"internal/meta/syntaxregistration#caseRegistrySlice",
		"internal/meta/syntaxregistration#registryOriginWalk.expression",
		"internal/meta/syntaxregistration#registryOriginWalk.helper",
		"internal/meta/syntaxregistration#registryOriginWalk.appended",
		"internal/meta/syntaxregistration#goSource.registryTailNeedsComma",
	}
	requireNamedCapacities(t, root, targets)
	t.Log("all8 changed production declarations for source-owned registry origins;unchanged75-line rendered helper cap;parent38 unchanged and retained;producer/model0")
}
