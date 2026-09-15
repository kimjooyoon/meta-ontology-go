package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func bindingDigest(surface Surface) string {
	var builder strings.Builder
	field(&builder, surface.SurfaceID.String())
	field(&builder, surface.CodeSymbolID.String())
	field(&builder, surface.SemanticOwnerID.String())
	field(&builder, surface.Binding.SourceMapID.String())
	return stableDigest(builder.String())
}
func normalizeRegistry(registry Registry, config Config) (map[semantic.ID]Surface, *evaluationIssue) {
	if registry.Schema == "" {
		return nil, required("registry")
	}
	if registry.Schema != RegistrySchemaV1 {
		return nil, failIssue(ReasonMalformedBinding, "registry schema")
	}
	if registry.Digest == "" {
		return nil, required("registry digest")
	}
	if issue := normalizeDigestValue(registry.Digest, "registry digest"); issue != nil {
		return nil, issue
	}
	if registry.Surfaces == nil {
		return nil, required("registry surfaces")
	}
	byID := make(map[semantic.ID]Surface, len(registry.Surfaces))
	bySymbol := make(map[semantic.ID]struct{}, len(registry.Surfaces))
	for _, raw := range registry.Surfaces {
		surface, issue := normalizeSurface(raw)
		if issue != nil {
			return nil, issue
		}
		if _, exists := byID[surface.SurfaceID]; exists {
			return nil, failIssue(ReasonDuplicateSurface, surface.SurfaceID.String())
		}
		if _, exists := bySymbol[surface.CodeSymbolID]; exists {
			return nil, failIssue(ReasonDuplicateSurface, surface.CodeSymbolID.String())
		}
		byID[surface.SurfaceID] = surface
		bySymbol[surface.CodeSymbolID] = struct{}{}
	}
	if stableDigest(registryCanonical(registry)) != registry.Digest {
		return nil, unknownIssue(ReasonStaleInput, "registry digest")
	}
	if registry.Digest != config.RegistryDigest {
		return nil, failIssue(ReasonDigestMismatch, "registry/config digest")
	}
	return byID, nil
}
