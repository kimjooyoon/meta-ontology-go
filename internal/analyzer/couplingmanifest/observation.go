package couplingmanifest

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeObservations(values []SourceMapObservation, inventory map[semantic.ID]Surface) (map[semantic.ID]SourceMapObservation, error) {
	result := make(map[semantic.ID]SourceMapObservation, len(values))
	seenSurfaces := make(map[semantic.ID]struct{}, len(values))
	seenSymbols := make(map[semantic.ID]struct{}, len(values))
	seenMaps := make(map[semantic.ID]struct{}, len(values))
	for _, value := range values {
		normalized, err := normalizeObservation(value)
		if err != nil {
			return nil, err
		}
		registered, ok := inventory[normalized.SurfaceID]
		if !ok {
			return nil, failError(CodeMalformedBinding, "surface ID %q is not registered", normalized.SurfaceID)
		}
		if normalized.CodeSymbolID != registered.CodeSymbolID || normalized.SemanticOwnerID != registered.SemanticOwnerID || normalized.SourceMapID != registered.Binding.SourceMapID {
			return nil, failError(CodeConflictingBinding, "surface ID %q conflicts with its registered identity tuple", normalized.SurfaceID)
		}
		if _, exists := seenSurfaces[normalized.SurfaceID]; exists {
			return nil, failError(CodeDuplicateBinding, "surface ID %q occurs more than once", normalized.SurfaceID)
		}
		if _, exists := seenSymbols[normalized.CodeSymbolID]; exists {
			return nil, failError(CodeConflictingBinding, "code symbol ID %q occurs more than once", normalized.CodeSymbolID)
		}
		if _, exists := seenMaps[normalized.SourceMapID]; exists {
			return nil, failError(CodeConflictingBinding, "source-map ID %q occurs more than once", normalized.SourceMapID)
		}
		result[normalized.SemanticOwnerID] = normalized
		seenSurfaces[normalized.SurfaceID] = struct{}{}
		seenSymbols[normalized.CodeSymbolID] = struct{}{}
		seenMaps[normalized.SourceMapID] = struct{}{}
	}
	return result, nil
}

func normalizeObservation(value SourceMapObservation) (SourceMapObservation, error) {
	for _, field := range []struct {
		value semantic.ID
		name  string
	}{
		{value.SurfaceID, "surface ID"}, {value.CodeSymbolID, "code symbol ID"},
		{value.SemanticOwnerID, "semantic owner ID"}, {value.SourceMapID, "source-map ID"},
	} {
		if _, err := normalizeID(field.value); err != nil {
			return SourceMapObservation{}, failError(CodeMalformedBinding, "%s: %v", field.name, err)
		}
	}
	if !validRole(value.Role) {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q has invalid role %q", value.SurfaceID, value.Role)
	}
	repoPath, err := normalizeRepoPath(value.Path)
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q path: %v", value.SurfaceID, err)
	}
	blobDigest, err := normalizeDigest(value.BlobDigest, "blob digest")
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	bindingDigest, err := normalizeDigest(value.BindingDigest, "binding digest")
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	sourceMapBindingDigest, err := normalizeDigest(value.SourceMapBindingDigest, "source-map binding digest")
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	value.Path, value.BlobDigest, value.BindingDigest, value.SourceMapBindingDigest = repoPath, blobDigest, bindingDigest, sourceMapBindingDigest
	return value, nil
}

func snapshotIndex(snapshot selectiveci.Snapshot) (map[semantic.ID]observed, error) {
	result := make(map[semantic.ID]observed)
	for _, source := range snapshot.Sources {
		blobDigest, err := normalizeDigest(source.BlobDigest, "snapshot blob digest")
		if err != nil {
			return nil, err
		}
		for _, binding := range source.Bindings {
			ownerID, err := normalizeIDString(binding.ID)
			if err != nil {
				return nil, err
			}
			bindingDigest, err := normalizeDigest(binding.BindingDigest, "snapshot binding digest")
			if err != nil {
				return nil, err
			}
			if _, exists := result[ownerID]; exists {
				return nil, fmt.Errorf("snapshot semantic owner ID %q occurs more than once", ownerID)
			}
			result[ownerID] = observed{Role: binding.Role, Path: source.Path, BlobDigest: blobDigest, BindingDigest: bindingDigest}
		}
	}
	return result, nil
}

func rejectUnregistered(values map[semantic.ID]observed, inventory map[semantic.ID]Surface) error {
	knownOwners := make(map[semantic.ID]struct{}, len(inventory))
	for _, surface := range inventory {
		knownOwners[surface.SemanticOwnerID] = struct{}{}
	}
	for ownerID := range values {
		if _, ok := knownOwners[ownerID]; !ok {
			return failError(CodeMalformedBinding, "semantic owner ID %q is not registered", ownerID)
		}
	}
	return nil
}

func resolveSide(snapshot map[semantic.ID]observed, bindings map[semantic.ID]SourceMapObservation, inventory map[semantic.ID]Surface, current bool) (map[semantic.ID]resolved, error) {
	result := make(map[semantic.ID]resolved, len(snapshot))
	for ownerID, value := range snapshot {
		binding, ok := bindings[ownerID]
		if !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has no exact source-map binding", ownerID)
		}
		registered := inventory[binding.SurfaceID]
		if registered.SemanticOwnerID != ownerID || binding.Path != value.Path || binding.Role != value.Role || binding.BlobDigest != value.BlobDigest || binding.BindingDigest != value.BindingDigest {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has a stale or conflicting source-map binding", ownerID)
		}
		if current && binding.SourceMapBindingDigest != registered.Binding.BindingDigest {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has a stale current source-map binding", ownerID)
		}
		result[ownerID] = resolved{Surface: registered, Observed: value}
	}
	return result, nil
}

func rejectUnobservedBindings(snapshot map[semantic.ID]observed, bindings map[semantic.ID]SourceMapObservation) error {
	for ownerID := range bindings {
		if _, observed := snapshot[ownerID]; !observed {
			return unknownError(CodeUnknownChangedSurface, "source-map binding for semantic owner ID %q has no matching snapshot observation", ownerID)
		}
	}
	return nil
}

func requireInventoryCoverage(inventory map[semantic.ID]Surface, before, head map[semantic.ID]resolved) error {
	for _, surface := range inventory {
		if _, beforeOK := before[surface.SemanticOwnerID]; !beforeOK {
			if _, headOK := head[surface.SemanticOwnerID]; !headOK {
				return unknownError(CodeUnknownChangedSurface, "registered surface %q has no before or head observation", surface.SurfaceID)
			}
		}
	}
	return nil
}
