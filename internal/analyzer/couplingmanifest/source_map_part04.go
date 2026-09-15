package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
