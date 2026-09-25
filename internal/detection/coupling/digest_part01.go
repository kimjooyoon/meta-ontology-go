package coupling

import (
	"sort"
	"strconv"
	"strings"
)

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' && char < 'a' || char > 'f' {
			return false
		}
	}
	return true
}
func field(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}
func registryCanonical(registry Registry) string {
	surfaces := append([]Surface(nil), registry.Surfaces...)
	sort.Slice(surfaces, func(i, j int) bool { return surfaces[i].SurfaceID < surfaces[j].SurfaceID })
	var builder strings.Builder
	field(&builder, RegistrySchemaV1)
	for _, surface := range surfaces {
		field(&builder, surface.SurfaceID.String())
		field(&builder, surface.CodeSymbolID.String())
		field(&builder, surface.SemanticOwnerID.String())
		field(&builder, surface.Binding.SourceMapID.String())
		field(&builder, surface.Binding.BindingDigest)
	}
	return builder.String()
}
func manifestCanonical(manifest ChangeManifest) string {
	entries := append([]ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var builder strings.Builder
	field(&builder, ManifestSchemaV1)
	field(&builder, strconv.FormatBool(manifest.Complete))
	field(&builder, strconv.FormatBool(manifest.ZeroChange))
	field(&builder, manifest.RegistryDigest)
	field(&builder, manifest.ToolchainDigest)
	field(&builder, manifest.ProfileDigest)
	field(&builder, manifest.BeforeSnapshotDigest)
	field(&builder, manifest.AfterSnapshotDigest)
	for _, entry := range entries {
		field(&builder, entry.SurfaceID.String())
		field(&builder, entry.CodeSymbolID.String())
		field(&builder, entry.SemanticOwnerID.String())
		field(&builder, entry.BeforeBindingDigest)
		field(&builder, entry.AfterBindingDigest)
		field(&builder, entry.BeforeBlobDigest)
		field(&builder, entry.AfterBlobDigest)
	}
	return builder.String()
}
