//go:build detector_bridge

package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"sort"
	"strconv"
	"strings"
)

func bridgeRegistryDigest(registry production.Registry) string {
	surfaces := append([]production.Surface(nil), registry.Surfaces...)
	sort.Slice(surfaces, func(i, j int) bool { return surfaces[i].SurfaceID < surfaces[j].SurfaceID })
	var b strings.Builder
	bridgeField(&b, production.RegistrySchemaV1)
	for _, s := range surfaces {
		bridgeField(&b, s.SurfaceID.String())
		bridgeField(&b, s.CodeSymbolID.String())
		bridgeField(&b, s.SemanticOwnerID.String())
		bridgeField(&b, s.Binding.SourceMapID.String())
		bridgeField(&b, s.Binding.BindingDigest)
	}
	return bridgeHash(b.String())
}
func bridgeManifestDigest(manifest production.ChangeManifest) string {
	entries := append([]production.ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var b strings.Builder
	bridgeField(&b, production.ManifestSchemaV1)
	bridgeField(&b, strconv.FormatBool(manifest.Complete))
	bridgeField(&b, strconv.FormatBool(manifest.ZeroChange))
	bridgeField(&b, manifest.RegistryDigest)
	bridgeField(&b, manifest.ToolchainDigest)
	bridgeField(&b, manifest.ProfileDigest)
	bridgeField(&b, manifest.BeforeSnapshotDigest)
	bridgeField(&b, manifest.AfterSnapshotDigest)
	for _, e := range entries {
		bridgeField(&b, e.SurfaceID.String())
		bridgeField(&b, e.CodeSymbolID.String())
		bridgeField(&b, e.SemanticOwnerID.String())
		bridgeField(&b, e.BeforeBindingDigest)
		bridgeField(&b, e.AfterBindingDigest)
		bridgeField(&b, e.BeforeBlobDigest)
		bridgeField(&b, e.AfterBlobDigest)
	}
	return bridgeHash(b.String())
}
func bridgeBaselineDigest(b production.BaselineConfig) string {
	var s strings.Builder
	bridgeField(&s, production.BaselineSchemaV1)
	bridgeField(&s, strconv.FormatBool(b.FullSuiteRequired))
	return bridgeHash(s.String())
}
func bridgeExternalDigest(r production.ExternalResourceReceipt) string {
	var b strings.Builder
	bridgeField(&b, production.ResourceSchemaV1)
	bridgeField(&b, r.SnapshotDigest)
	bridgeField(&b, r.ProviderDigest)
	bridgeField(&b, r.ObserverDigest)
	if r.CPUWorkUnits != nil {
		bridgeField(&b, "cpu_work_units")
		bridgeField(&b, strconv.FormatUint(*r.CPUWorkUnits, 10))
	}
	if r.PeakMemoryBytes != nil {
		bridgeField(&b, "peak_memory_bytes")
		bridgeField(&b, strconv.FormatUint(*r.PeakMemoryBytes, 10))
	}
	if r.DeterministicWorkUnits != nil {
		bridgeField(&b, "deterministic_work_units")
		bridgeField(&b, strconv.FormatUint(*r.DeterministicWorkUnits, 10))
	}
	return bridgeHash(b.String())
}
