package couplingmanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strconv"
	"strings"
)

func testObservations(snapshot selectiveci.Snapshot, surfaces []detector.Surface) []SourceMapObservation {
	byOwner := make(map[semantic.ID]detector.Surface, len(surfaces))
	for _, surface := range surfaces {
		byOwner[surface.SemanticOwnerID] = surface
	}
	result := make([]SourceMapObservation, 0)
	for _, source := range snapshot.Sources {
		for _, binding := range source.Bindings {
			owner := semantic.MustIdentity(binding.ID)
			surface, ok := byOwner[owner]
			if !ok {
				continue
			}
			result = append(result, SourceMapObservation{
				SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID,
				SourceMapID: surface.Binding.SourceMapID, Role: binding.Role, Path: source.Path,
				BlobDigest: source.BlobDigest, BindingDigest: binding.BindingDigest,
				SourceMapBindingDigest: surface.Binding.BindingDigest,
			})
		}
	}
	return result
}
func testBaseline() detector.BaselineConfig {
	baseline := detector.BaselineConfig{Schema: detector.BaselineSchemaV1, FullSuiteRequired: true}
	baseline.Digest = stableDigest(testFields(detector.BaselineSchemaV1, strconv.FormatBool(true)))
	return baseline
}
func bindingDigestForTest(surface detector.Surface) string {
	return stableDigest(testFields(surface.SurfaceID.String(), surface.CodeSymbolID.String(), surface.SemanticOwnerID.String(), surface.Binding.SourceMapID.String()))
}
func registryDigestForTest(registry detector.Registry) string {
	surfaces := append([]detector.Surface(nil), registry.Surfaces...)
	sort.Slice(surfaces, func(i, j int) bool { return surfaces[i].SurfaceID < surfaces[j].SurfaceID })
	values := []string{detector.RegistrySchemaV1}
	for _, surface := range surfaces {
		values = append(values, surface.SurfaceID.String(), surface.CodeSymbolID.String(), surface.SemanticOwnerID.String(), surface.Binding.SourceMapID.String(), surface.Binding.BindingDigest)
	}
	return stableDigest(testFields(values...))
}
func testFields(values ...string) string {
	var builder strings.Builder
	for _, value := range values {
		builder.WriteString(strconv.Itoa(len(value)))
		builder.WriteByte(':')
		builder.WriteString(value)
		builder.WriteByte('|')
	}
	return builder.String()
}
func testDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}
func rawTestDigest(value string) string { return strings.TrimPrefix(value, "sha256:") }
