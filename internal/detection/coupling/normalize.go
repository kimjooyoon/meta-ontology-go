package coupling

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type evaluationIssue struct {
	status Status
	code   ReasonCode
	detail string
}

func failIssue(code ReasonCode, detail string) *evaluationIssue {
	return &evaluationIssue{status: StatusFailClosed, code: code, detail: detail}
}

func unknownIssue(code ReasonCode, detail string) *evaluationIssue {
	return &evaluationIssue{status: StatusUnknown, code: code, detail: detail}
}

func required(fieldName string) *evaluationIssue {
	return unknownIssue(ReasonRequiredInputMissing, fieldName)
}

func normalizeID(id semantic.ID, fieldName string) (semantic.ID, *evaluationIssue) {
	parsed, err := semantic.ParseIdentity(id.String())
	if err != nil || parsed != id {
		return "", failIssue(ReasonMalformedBinding, fieldName)
	}
	return parsed, nil
}

func normalizeDigestValue(value, fieldName string) *evaluationIssue {
	if value == "" {
		return required(fieldName)
	}
	if !validDigest(value) {
		return failIssue(ReasonMalformedBinding, fieldName)
	}
	return nil
}

func normalizeConfig(config Config) *evaluationIssue {
	if config.Schema == "" || config.Baseline.Schema == "" {
		return required("config")
	}
	if config.Schema != ConfigSchemaV1 || config.Baseline.Schema != BaselineSchemaV1 {
		return failIssue(ReasonMalformedBinding, "config schema")
	}
	if !config.Baseline.FullSuiteRequired {
		return failIssue(ReasonMalformedBinding, "full-suite baseline is not enabled")
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{config.RegistryDigest, "config registry digest"},
		{config.ToolchainDigest, "config toolchain digest"},
		{config.ProfileDigest, "config profile digest"},
		{config.SnapshotDigest, "config snapshot digest"},
		{config.ExpectedProviderDigest, "expected provider digest"},
		{config.ExpectedObserverDigest, "expected observer digest"},
		{config.Baseline.Digest, "baseline digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if stableDigest(baselineCanonical(config.Baseline)) != config.Baseline.Digest {
		return unknownIssue(ReasonStaleInput, "baseline digest")
	}
	return nil
}

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

func normalizeSurface(raw Surface) (Surface, *evaluationIssue) {
	surface := raw
	var issue *evaluationIssue
	if surface.SurfaceID, issue = normalizeID(raw.SurfaceID, "surface ID"); issue != nil {
		return Surface{}, issue
	}
	if surface.CodeSymbolID, issue = normalizeID(raw.CodeSymbolID, "code symbol ID"); issue != nil {
		return Surface{}, issue
	}
	if surface.SemanticOwnerID, issue = normalizeID(raw.SemanticOwnerID, "semantic owner ID"); issue != nil {
		return Surface{}, issue
	}
	if surface.Binding.SourceMapID, issue = normalizeID(raw.Binding.SourceMapID, "source map ID"); issue != nil {
		return Surface{}, issue
	}
	if issue = normalizeDigestValue(surface.Binding.BindingDigest, "source map binding digest"); issue != nil {
		return Surface{}, issue
	}
	if expected := bindingDigest(surface); expected != surface.Binding.BindingDigest {
		return Surface{}, failIssue(ReasonSourceMapMismatch, surface.SurfaceID.String())
	}
	return surface, nil
}

func normalizeManifest(
	manifest ChangeManifest, config Config, registry map[semantic.ID]Surface,
) ([]ManifestEntry, *evaluationIssue) {
	if manifest.Schema == "" {
		return nil, required("manifest")
	}
	if manifest.Schema != ManifestSchemaV1 {
		return nil, failIssue(ReasonMalformedBinding, "manifest schema")
	}
	if !manifest.Complete {
		return nil, required("complete source-backed change manifest")
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{manifest.RegistryDigest, "manifest registry digest"},
		{manifest.ToolchainDigest, "manifest toolchain digest"},
		{manifest.ProfileDigest, "manifest profile digest"},
		{manifest.BeforeSnapshotDigest, "manifest before snapshot digest"},
		{manifest.AfterSnapshotDigest, "manifest after snapshot digest"},
		{manifest.Digest, "manifest digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return nil, issue
		}
	}
	if manifest.RegistryDigest != config.RegistryDigest || manifest.ToolchainDigest != config.ToolchainDigest ||
		manifest.ProfileDigest != config.ProfileDigest || manifest.AfterSnapshotDigest != config.SnapshotDigest {
		return nil, failIssue(ReasonDigestMismatch, "manifest/config digest")
	}
	if stableDigest(manifestCanonical(manifest)) != manifest.Digest {
		return nil, unknownIssue(ReasonStaleInput, "manifest digest")
	}
	if manifest.Entries == nil {
		return nil, required("complete manifest entries")
	}
	if len(manifest.Entries) != len(registry) {
		return nil, failIssue(ReasonMalformedBinding, "manifest entry cardinality")
	}
	entries := append([]ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	seen := make(map[semantic.ID]struct{}, len(entries))
	for index := range entries {
		entry := &entries[index]
		if _, duplicate := seen[entry.SurfaceID]; duplicate {
			return nil, failIssue(ReasonDuplicateSurface, entry.SurfaceID.String())
		}
		seen[entry.SurfaceID] = struct{}{}
		surface, exists := registry[entry.SurfaceID]
		if !exists {
			return nil, failIssue(ReasonSurfaceUnregistered, entry.SurfaceID.String())
		}
		if entry.CodeSymbolID != surface.CodeSymbolID || entry.SemanticOwnerID != surface.SemanticOwnerID {
			return nil, failIssue(ReasonSourceMapMismatch, entry.SurfaceID.String())
		}
		if entry.AfterBindingDigest != surface.Binding.BindingDigest {
			return nil, failIssue(ReasonSourceMapMismatch, entry.SurfaceID.String())
		}
		if entry.BeforeSourcePath == "" || entry.AfterSourcePath == "" {
			return nil, required("manifest source path")
		}
		for _, value := range []struct {
			value string
			name  string
		}{
			{entry.BeforeBindingDigest, "before binding digest"},
			{entry.AfterBindingDigest, "after binding digest"},
			{entry.BeforeBlobDigest, "before blob digest"},
			{entry.AfterBlobDigest, "after blob digest"},
		} {
			if issue := normalizeDigestValue(value.value, value.name); issue != nil {
				return nil, issue
			}
		}
	}
	changed := make([]ManifestEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.BeforeBindingDigest != entry.AfterBindingDigest || entry.BeforeBlobDigest != entry.AfterBlobDigest {
			changed = append(changed, entry)
		}
	}
	if manifest.ZeroChange != (len(changed) == 0) {
		return nil, failIssue(ReasonContradictoryReceipt, "zero-change claim")
	}
	return changed, nil
}

func sortedIDs(values []semantic.ID) []semantic.ID {
	result := append([]semantic.ID(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func sortedReasons(values []Reason) []Reason {
	result := append([]Reason(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Code != result[j].Code {
			return result[i].Code < result[j].Code
		}
		return result[i].Detail < result[j].Detail
	})
	return result
}

func detailf(format string, args ...any) string { return fmt.Sprintf(format, args...) }
