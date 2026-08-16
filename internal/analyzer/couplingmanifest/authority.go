package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeAuthority(input RegistrySourceMap) (normalizedAuthority, error) {
	if input.Schema == "" || input.RegistryDigest == "" || input.SourceMapDigest == "" || input.ToolchainDigest == "" || input.ProfileDigest == "" || input.Inventory == nil || input.Before == nil || input.Head == nil {
		return normalizedAuthority{}, unknownError(CodeMissingAuthority, "versioned registry/source-map, toolchain/profile, inventory, and both side bindings are required")
	}
	if input.Schema != AuthoritySchemaV1 {
		return normalizedAuthority{}, failError(CodeInvalidSchema, "authority schema %q is not %q", input.Schema, AuthoritySchemaV1)
	}
	registryDigest, err := normalizeDigest(input.RegistryDigest, "registry digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	sourceMapDigest, err := normalizeDigest(input.SourceMapDigest, "source-map digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	toolchainDigest, err := normalizeDigest(input.ToolchainDigest, "toolchain digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	profileDigest, err := normalizeDigest(input.ProfileDigest, "profile digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	if len(input.CandidateBindings) != 0 {
		return normalizedAuthority{}, unknownError(CodeCandidateBinding, "candidate bindings cannot become authoritative")
	}
	if len(input.DerivedBindings) != 0 {
		return normalizedAuthority{}, unknownError(CodeDerivedBinding, "derived bindings cannot become authoritative")
	}
	inventory, err := normalizeInventory(input.Inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	before, err := normalizeObservations(input.Before, inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	head, err := normalizeObservations(input.Head, inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	return normalizedAuthority{registryDigest: registryDigest, sourceMapDigest: sourceMapDigest, toolchainDigest: toolchainDigest, profileDigest: profileDigest, inventory: inventory, before: before, head: head}, nil
}

func normalizeInventory(values []Surface) (map[semantic.ID]Surface, error) {
	result := make(map[semantic.ID]Surface, len(values))
	seenSymbols := make(map[semantic.ID]semantic.ID, len(values))
	seenOwners := make(map[semantic.ID]semantic.ID, len(values))
	seenMaps := make(map[semantic.ID]semantic.ID, len(values))
	for _, value := range values {
		normalized, err := normalizeSurface(value)
		if err != nil {
			return nil, err
		}
		if _, exists := result[normalized.SurfaceID]; exists {
			return nil, failError(CodeDuplicateBinding, "surface ID %q occurs more than once", normalized.SurfaceID)
		}
		if previous, exists := seenSymbols[normalized.CodeSymbolID]; exists {
			return nil, failError(CodeConflictingBinding, "code symbol ID %q resolves to both %q and %q", normalized.CodeSymbolID, previous, normalized.SurfaceID)
		}
		if previous, exists := seenOwners[normalized.SemanticOwnerID]; exists {
			return nil, failError(CodeConflictingBinding, "semantic owner ID %q resolves to both %q and %q", normalized.SemanticOwnerID, previous, normalized.SurfaceID)
		}
		if previous, exists := seenMaps[normalized.Binding.SourceMapID]; exists {
			return nil, failError(CodeConflictingBinding, "source-map ID %q resolves to both %q and %q", normalized.Binding.SourceMapID, previous, normalized.SurfaceID)
		}
		result[normalized.SurfaceID] = normalized
		seenSymbols[normalized.CodeSymbolID] = normalized.SurfaceID
		seenOwners[normalized.SemanticOwnerID] = normalized.SurfaceID
		seenMaps[normalized.Binding.SourceMapID] = normalized.SurfaceID
	}
	return result, nil
}

func normalizeSurface(value Surface) (Surface, error) {
	for _, field := range []struct {
		value semantic.ID
		name  string
	}{
		{value.SurfaceID, "surface ID"}, {value.CodeSymbolID, "code symbol ID"},
		{value.SemanticOwnerID, "semantic owner ID"}, {value.Binding.SourceMapID, "source-map ID"},
	} {
		if _, err := normalizeID(field.value); err != nil {
			return Surface{}, failError(CodeMalformedBinding, "%s: %v", field.name, err)
		}
	}
	bindingDigest, err := normalizeDigest(value.Binding.BindingDigest, "source-map binding digest")
	if err != nil {
		return Surface{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	value.Binding.BindingDigest = bindingDigest
	if expected := sourceMapBindingDigest(value); expected != bindingDigest {
		return Surface{}, failError(CodeConflictingBinding, "surface %q has a stale source-map binding digest", value.SurfaceID)
	}
	return value, nil
}

func validRole(role semanticbinding.Role) bool {
	return role == semanticbinding.RoleHandwrittenImpl || role == semanticbinding.RoleGeneratedImpl || role == semanticbinding.RoleAdapter
}
