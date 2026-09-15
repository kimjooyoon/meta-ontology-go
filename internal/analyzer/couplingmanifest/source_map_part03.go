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
		if err := assertSourceMapObservationFields(value); err != nil {
			if classified, ok := err.(*ConstructionError); ok {
				return nil, classified
			}
			return nil, unknownError(CodeUnknownChangedSurface, "%v", err)
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
		prepared, err := parseSourceMapObservationDigests(value)
		if err != nil {
			return nil, err
		}
		value = prepared
		if err := validateSourceMapObservationSurface(registered, value, seenSymbols, seenMaps); err != nil {
			return nil, err
		}
		if _, exists := result[value.SemanticOwnerID]; exists {
			return nil, failError(CodeDuplicateBinding, "semantic owner ID %q occurs more than once", value.SemanticOwnerID)
		}
		result[value.SemanticOwnerID] = value
		seenSymbols[value.CodeSymbolID] = struct{}{}
		seenMaps[value.SourceMapID] = struct{}{}
	}
	return result, nil
}
