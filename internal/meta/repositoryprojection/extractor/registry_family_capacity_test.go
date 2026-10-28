package extractor

import (
	"os"
	"testing"
)

func TestRegistryFamilySplitRenderedCapacity(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{
		"internal/meta/syntaxregistration#generateRegistry",
		"internal/meta/syntaxregistration#goSource.appendBoundedRegistryFamily",
		"internal/meta/syntaxregistration#goSource.registryLiteralOwner",
		"internal/meta/syntaxregistration#goSource.renderRegistryFamily",
		"internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup5",
		"internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup5Part0",
		"internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup5Part1",
	}
	requireNamedCapacities(t, root, targets)
}
