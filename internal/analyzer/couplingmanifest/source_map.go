package couplingmanifest

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateSourceMapContext(input Input) *ConstructionError {
	if input.SourceMap.Before == nil || input.SourceMap.Head == nil {
		return unknownError(CodeMissingAuthority, "before and head source-map observations are required")
	}
	if len(input.SourceMap.CandidateBindings) != 0 {
		return unknownError(CodeCandidateBinding, "candidate source-map observations are not authoritative")
	}
	if len(input.SourceMap.DerivedBindings) != 0 {
		return unknownError(CodeDerivedBinding, "derived source-map observations are not authoritative")
	}
	if _, err := registrySurfaceIndex(input.Authority.Registry.Surfaces); err != nil {
		return constructionError(err)
	}
	return nil
}

func resolveSnapshots(input Input) (map[semantic.ID]resolved, map[semantic.ID]resolved, *ConstructionError) {
	surfaces, err := registrySurfaceIndex(input.Authority.Registry.Surfaces)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	beforeSnapshot, err := snapshotIndex(*input.Before)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	headSnapshot, err := snapshotIndex(*input.Head)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnregistered(beforeSnapshot, surfaces); err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnregistered(headSnapshot, surfaces); err != nil {
		return nil, nil, constructionError(err)
	}
	beforeBindings, err := observationIndex(input.SourceMap.Before, surfaces)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	headBindings, err := observationIndex(input.SourceMap.Head, surfaces)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnobservedBindings(beforeSnapshot, beforeBindings); err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnobservedBindings(headSnapshot, headBindings); err != nil {
		return nil, nil, constructionError(err)
	}
	before, err := resolveSide(beforeSnapshot, beforeBindings, surfaces, false)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	head, err := resolveSide(headSnapshot, headBindings, surfaces, true)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	if err := requireCoverage(surfaces, before, head); err != nil {
		return nil, nil, constructionError(err)
	}
	return before, head, nil
}

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

func snapshotIndex(snapshot selectiveci.Snapshot) (map[semantic.ID]observed, error) {
	result := make(map[semantic.ID]observed)
	for _, source := range snapshot.Sources {
		blobDigest, err := rawDigest(source.BlobDigest)
		if err != nil {
			return nil, err
		}
		for _, binding := range source.Bindings {
			if binding.ID == "" {
				return nil, unknownError(CodeUnknownChangedSurface, "snapshot semantic owner ID is missing")
			}
			ownerID, err := canonicalIDString(binding.ID)
			if err != nil {
				return nil, failError(CodeMalformedBinding, "snapshot semantic owner ID: %v", err)
			}
			bindingDigest, err := rawDigest(binding.BindingDigest)
			if err != nil {
				return nil, err
			}
			if _, exists := result[ownerID]; exists {
				return nil, failError(CodeDuplicateBinding, "snapshot semantic owner ID %q occurs more than once", ownerID)
			}
			result[ownerID] = observed{Role: string(binding.Role), Path: source.Path, BlobDigest: blobDigest, BindingDigest: bindingDigest}
		}
	}
	return result, nil
}

func rejectUnregistered(values map[semantic.ID]observed, surfaces map[semantic.ID]detector.Surface) error {
	owners := make(map[semantic.ID]struct{}, len(surfaces))
	for _, surface := range surfaces {
		owners[surface.SemanticOwnerID] = struct{}{}
	}
	for owner := range values {
		if _, ok := owners[owner]; !ok {
			return unknownError(CodeUnknownChangedSurface, "semantic owner ID %q is not registered", owner)
		}
	}
	return nil
}

func rejectUnobservedBindings(snapshot map[semantic.ID]observed, bindings map[semantic.ID]SourceMapObservation) error {
	for owner := range bindings {
		if _, ok := snapshot[owner]; !ok {
			return unknownError(CodeUnknownChangedSurface, "source-map observation has no matching snapshot binding")
		}
	}
	return nil
}

func resolveSide(snapshot map[semantic.ID]observed, bindings map[semantic.ID]SourceMapObservation, surfaces map[semantic.ID]detector.Surface, current bool) (map[semantic.ID]resolved, error) {
	result := make(map[semantic.ID]resolved, len(snapshot))
	for owner, value := range snapshot {
		binding, ok := bindings[owner]
		if !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "snapshot binding has no exact source-map observation")
		}
		surface := surfaces[binding.SurfaceID]
		if value.Path != binding.Path || value.BlobDigest != binding.BlobDigest || value.BindingDigest != binding.BindingDigest || value.Role != string(binding.Role) {
			return nil, unknownError(CodeUnknownChangedSurface, "source-map observation is stale or conflicting")
		}
		if current && binding.SourceMapBindingDigest != surface.Binding.BindingDigest {
			return nil, unknownError(CodeUnknownChangedSurface, "current source-map binding is stale")
		}
		result[owner] = resolved{Observed: observed{Role: value.Role, Path: value.Path, BlobDigest: value.BlobDigest, BindingDigest: binding.SourceMapBindingDigest}}
	}
	return result, nil
}

func requireCoverage(surfaces map[semantic.ID]detector.Surface, before, head map[semantic.ID]resolved) error {
	for _, surface := range surfaces {
		if _, beforeOK := before[surface.SemanticOwnerID]; !beforeOK {
			if _, headOK := head[surface.SemanticOwnerID]; !headOK {
				return unknownError(CodeUnknownChangedSurface, "registered surface has no before or head observation")
			}
		}
	}
	return nil
}

func sortedSurfaces(values []detector.Surface) []detector.Surface {
	result := append([]detector.Surface(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].SurfaceID < result[j].SurfaceID })
	return result
}
