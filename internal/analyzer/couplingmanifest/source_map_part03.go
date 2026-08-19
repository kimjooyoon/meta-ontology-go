package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func observationIndex(values []SourceMapObservation, surfaces map[semantic.ID]detector.Surface) (map[semantic.ID]SourceMapObservation, error) {
	result := make(map[semantic.ID]SourceMapObservation, len(values))
	seenSymbols := make(map[semantic.ID]struct{}, len(values))
	seenMaps := make(map[semantic.ID]struct{}, len(values))
	registeredSymbols := make(map[semantic.ID]struct{}, len(surfaces))
	registeredOwners := make(map[semantic.ID]struct{}, len(surfaces))
	registeredMaps := make(map[semantic.ID]struct{}, len(surfaces))
	for _, surface := range surfaces {
		registeredSymbols[surface.CodeSymbolID] = struct{}{}
		registeredOwners[surface.SemanticOwnerID] = struct{}{}
		registeredMaps[surface.Binding.SourceMapID] = struct{}{}
	}
	for _, value := range values {
		if value.SurfaceID == "" || value.CodeSymbolID == "" || value.SemanticOwnerID == "" || value.SourceMapID == "" {
			return nil, unknownError(CodeUnknownChangedSurface, "source-map observation identity is missing")
		}
		if _, err := canonicalID(value.SurfaceID); err != nil {
			return nil, failError(CodeMalformedBinding, "observation surface ID: %v", err)
		}
		if _, err := canonicalID(value.CodeSymbolID); err != nil {
			return nil, failError(CodeMalformedBinding, "observation code symbol ID: %v", err)
		}
		if _, err := canonicalID(value.SemanticOwnerID); err != nil {
			return nil, failError(CodeMalformedBinding, "observation semantic owner ID: %v", err)
		}
		if _, err := canonicalID(value.SourceMapID); err != nil {
			return nil, failError(CodeMalformedBinding, "observation source-map ID: %v", err)
		}
		registered, ok := surfaces[value.SurfaceID]
		if !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "surface ID %q is not registered", value.SurfaceID)
		}
		if _, ok := registeredSymbols[value.CodeSymbolID]; !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "code symbol ID %q is not registered", value.CodeSymbolID)
		}
		if _, ok := registeredOwners[value.SemanticOwnerID]; !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q is not registered", value.SemanticOwnerID)
		}
		if _, ok := registeredMaps[value.SourceMapID]; !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "source-map ID %q is not registered", value.SourceMapID)
		}
		if value.CodeSymbolID != registered.CodeSymbolID || value.SemanticOwnerID != registered.SemanticOwnerID || value.SourceMapID != registered.Binding.SourceMapID {
			return nil, failError(CodeConflictingBinding, "observation identity tuple differs from detector registry")
		}
		blobDigest, blobErr := rawDigest(value.BlobDigest)
		bindingDigest, bindingErr := rawDigest(value.BindingDigest)
		sourceMapBindingDigest, sourceMapBindingErr := rawDigest(value.SourceMapBindingDigest)
		if blobErr != nil || bindingErr != nil || sourceMapBindingErr != nil {
			return nil, failError(CodeMalformedBinding, "source-map observation digest is malformed")
		}
		value.BlobDigest, value.BindingDigest, value.SourceMapBindingDigest = blobDigest, bindingDigest, sourceMapBindingDigest
		if value.SourceMapBindingDigest != registered.Binding.BindingDigest {
			return nil, failError(CodeConflictingBinding, "observation source-map binding differs from detector registry")
		}
		if _, exists := result[value.SemanticOwnerID]; exists {
			return nil, failError(CodeDuplicateBinding, "semantic owner ID %q occurs more than once", value.SemanticOwnerID)
		}
		if _, exists := seenSymbols[value.CodeSymbolID]; exists {
			return nil, failError(CodeConflictingBinding, "code symbol ID %q occurs more than once", value.CodeSymbolID)
		}
		if _, exists := seenMaps[value.SourceMapID]; exists {
			return nil, failError(CodeConflictingBinding, "source-map ID %q occurs more than once", value.SourceMapID)
		}
		result[value.SemanticOwnerID] = value
		seenSymbols[value.CodeSymbolID] = struct{}{}
		seenMaps[value.SourceMapID] = struct{}{}
	}
	return result, nil
}
