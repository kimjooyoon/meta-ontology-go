package extractor

import "testing"

func TestTextInputBoundaryDeclarationCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodyexecution#decodeCompositionTextData",
		"internal/bodyexecution#decodeTextScalar",
		"internal/bodyexecution#validateJSONTextEscapes",
		"internal/bodyexecution#jsonEscapeUnit",
		"internal/bodyexecution#canonicalScalar",
	})
}
