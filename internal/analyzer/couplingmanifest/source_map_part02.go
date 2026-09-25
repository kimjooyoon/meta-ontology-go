package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func registrySurfaceIndex(values []detector.Surface) (map[semantic.ID]detector.Surface, error) {
	result := make(map[semantic.ID]detector.Surface, len(values))
	seenSymbols := make(map[semantic.ID]struct{}, len(values))
	seenOwners := make(map[semantic.ID]struct{}, len(values))
	seenMaps := make(map[semantic.ID]struct{}, len(values))
	for _, surface := range values {
		if _, err := canonicalID(surface.SurfaceID); err != nil {
			return nil, failError(CodeMalformedBinding, "surface ID: %v", err)
		}
		if _, err := canonicalID(surface.CodeSymbolID); err != nil {
			return nil, failError(CodeMalformedBinding, "code symbol ID: %v", err)
		}
		if _, err := canonicalID(surface.SemanticOwnerID); err != nil {
			return nil, failError(CodeMalformedBinding, "semantic owner ID: %v", err)
		}
		if _, err := canonicalID(surface.Binding.SourceMapID); err != nil {
			return nil, failError(CodeMalformedBinding, "source-map ID: %v", err)
		}
		if _, err := rawDigest(surface.Binding.BindingDigest); err != nil {
			return nil, failError(CodeMalformedBinding, "source-map binding digest: %v", err)
		}
		if _, exists := result[surface.SurfaceID]; exists {
			return nil, failError(CodeDuplicateBinding, "surface ID %q occurs more than once", surface.SurfaceID)
		}
		if _, exists := seenSymbols[surface.CodeSymbolID]; exists {
			return nil, failError(CodeDuplicateBinding, "code symbol ID %q occurs more than once", surface.CodeSymbolID)
		}
		if _, exists := seenOwners[surface.SemanticOwnerID]; exists {
			return nil, failError(CodeDuplicateBinding, "semantic owner ID %q occurs more than once", surface.SemanticOwnerID)
		}
		if _, exists := seenMaps[surface.Binding.SourceMapID]; exists {
			return nil, failError(CodeDuplicateBinding, "source-map ID %q occurs more than once", surface.Binding.SourceMapID)
		}
		result[surface.SurfaceID] = surface
		seenSymbols[surface.CodeSymbolID] = struct{}{}
		seenOwners[surface.SemanticOwnerID] = struct{}{}
		seenMaps[surface.Binding.SourceMapID] = struct{}{}
	}
	return result, nil
}
