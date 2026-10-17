package extractor

import "testing"

func TestTextInputBoundaryDeclarationCapacity(t *testing.T) {
	// Follow the shared receiver/scanner after the host move, while retaining
	// the compatibility entry points in the original execution package.
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodyexecution#decodeCompositionTextData",
		"internal/bodyexecution#decodeTextScalar",
		"internal/bodyexecution#decodeBoundedTextScalar",
		"internal/bodyexecution#validateJSONTextEscapes",
		"internal/valuejson#DecodeText",
		"internal/valuejson#CheckEscapesInValidJSON",
		"internal/valuejson#escapeUnit",
		"internal/valuejson#Object",
		"internal/bodyexecution#canonicalScalar",
	})
}
