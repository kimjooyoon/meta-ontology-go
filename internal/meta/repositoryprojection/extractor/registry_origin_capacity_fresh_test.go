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
		"internal/meta/syntaxregistration#goSource.registryCaseLeaves",
		"internal/meta/syntaxregistration#goSource.registryFieldLeaves",
		"internal/meta/syntaxregistration#namedRegistryType",
		"internal/meta/syntaxregistration#registryOriginWalk.sliceType",
		"internal/meta/syntaxregistration#registryOriginWalk.expression",
		"internal/meta/syntaxregistration#registryOriginWalk.helper",
		"internal/meta/syntaxregistration#registryOriginWalk.appended",
		"internal/meta/syntaxregistration#goSource.registryTailNeedsComma",
		"internal/meta/syntaxregistration#bindNativeCaseSources",
		"internal/meta/syntaxregistration#readCaseSourceLiteral",
	}
	requireNamedCapacities(t, root, targets)
	t.Logf("%d required registry-origin declarations; unchanged 75-line rendered helper cap; producer/model 0", len(targets))
}
