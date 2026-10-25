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
		"internal/meta/syntaxregistration#goSource.registryFieldOrigins",
		"internal/meta/syntaxregistration#registryOrigins.literal",
		"internal/meta/syntaxregistration#registryReturnLiteral",
		"internal/meta/syntaxregistration#registryLiteralField",
		"internal/meta/syntaxregistration#namedRegistryType",
		"internal/meta/syntaxregistration#registryOriginWalk.sliceType",
		"internal/meta/syntaxregistration#registryOriginWalk.expression",
		"internal/meta/syntaxregistration#registryOriginWalk.helper",
		"internal/meta/syntaxregistration#registryOriginWalk.appended",
		"internal/meta/syntaxregistration#registryOriginWalk.localElements",
		"internal/meta/syntaxregistration#goSource.registryTailNeedsComma",
		"internal/meta/syntaxregistration#bindNativeCaseSources",
		"internal/meta/syntaxregistration#readCaseSourceLiteral",
		"internal/meta/syntaxregistration#readNativeCaseSources",
		"internal/meta/syntaxregistration#generatePartitionNative",
		"internal/meta/syntaxregistration#bindPartitionNative",
		"internal/meta/syntaxregistration#bindPartitionRegistryRemainder",
		"internal/meta/syntaxregistration#readPartitionRecord",
		"internal/meta/syntaxregistration#editPartitionFamilies",
		"internal/meta/syntaxregistration#editPartitionPackages",
		"internal/meta/syntaxregistration#readPartitionNamedInputs",
		"internal/meta/syntaxregistration#capturePartitionSources",
		"internal/meta/syntaxregistration#PreparePartitionRegistrationCandidate",
	}
	requireNamedCapacities(t, root, targets)
	t.Logf("%d required registry-origin declarations; unchanged 75-line rendered helper cap; producer/model 0", len(targets))
}
