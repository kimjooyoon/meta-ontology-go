package couplingmanifest

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func assertSourceMapObservationFields(value SourceMapObservation) error {
	if value.SurfaceID == "" || value.CodeSymbolID == "" || value.SemanticOwnerID == "" || value.SourceMapID == "" {
		return fmt.Errorf("source-map observation identity is missing")
	}
	if _, err := canonicalID(value.SurfaceID); err != nil {
		return failError(CodeMalformedBinding, "observation surface ID: %v", err)
	}
	if _, err := canonicalID(value.CodeSymbolID); err != nil {
		return failError(CodeMalformedBinding, "observation code symbol ID: %v", err)
	}
	if _, err := canonicalID(value.SemanticOwnerID); err != nil {
		return failError(CodeMalformedBinding, "observation semantic owner ID: %v", err)
	}
	if _, err := canonicalID(value.SourceMapID); err != nil {
		return failError(CodeMalformedBinding, "observation source-map ID: %v", err)
	}
	return nil
}

func parseSourceMapObservationDigests(value SourceMapObservation) (SourceMapObservation, error) {
	blobDigest, blobErr := rawDigest(value.BlobDigest)
	bindingDigest, bindingErr := rawDigest(value.BindingDigest)
	sourceMapBindingDigest, sourceMapBindingErr := rawDigest(value.SourceMapBindingDigest)
	if blobErr != nil || bindingErr != nil || sourceMapBindingErr != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "source-map observation digest is malformed")
	}
	value.BlobDigest, value.BindingDigest, value.SourceMapBindingDigest = blobDigest, bindingDigest, sourceMapBindingDigest
	return value, nil
}

func validateSourceMapObservationSurface(registered detector.Surface, value SourceMapObservation, seenSymbols, seenMaps map[semantic.ID]struct{}) error {
	if _, ok := seenSymbols[value.CodeSymbolID]; ok {
		return failError(CodeConflictingBinding, "code symbol ID %q occurs more than once", value.CodeSymbolID)
	}
	if _, ok := seenMaps[value.SourceMapID]; ok {
		return failError(CodeConflictingBinding, "source-map ID %q occurs more than once", value.SourceMapID)
	}
	if value.SourceMapBindingDigest != registered.Binding.BindingDigest {
		return failError(CodeConflictingBinding, "observation source-map binding differs from detector registry")
	}
	return nil
}
