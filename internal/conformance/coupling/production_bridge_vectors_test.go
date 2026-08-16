//go:build detector_bridge

package coupling

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func productionInputSnapshot(input production.Input) string {
	data, err := json.Marshal(input)
	if err != nil {
		panic("cannot snapshot producer input: " + err.Error())
	}
	return string(data)
}

func productionVectorFromResult(result production.Result, input production.Input, authority production.AuthorityContext) productionVector {
	accepted := make([]string, 0, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		accepted = append(accepted, id.String())
	}
	sort.Strings(accepted)
	return productionVector{Schema: result.Schema, Decision: bridgeDecision(result.Status), Reasons: append([]production.Reason(nil), result.Reasons...), Accepted: accepted, Observation: result.Observation, FullSuite: result.FullSuiteRequired, InputDigest: result.InputDigest, ResultDigest: result.Digest, Bindings: productionBindings(input, authority)}
}

func productionBindings(input production.Input, authority production.AuthorityContext) productionBindingVector {
	b := productionBindingVector{AuthoritySchema: authority.Schema, AuthorityRegistryDigest: authority.Registry.Digest, AuthorityToolchainDigest: authority.ToolchainDigest, AuthorityProfileDigest: authority.ProfileDigest, AuthoritySnapshotDigest: authority.SnapshotDigest, AuthorityProviderDigest: authority.ExpectedProviderDigest, AuthorityObserverDigest: authority.ExpectedObserverDigest, AuthorityBaselineDigest: authority.Baseline.Digest, AuthorityExternalReceiptRequired: authority.ExternalReceiptRequired, PacketRegistryDigest: input.Registry.Digest, ConfigRegistryDigest: input.Config.RegistryDigest, ConfigToolchainDigest: input.Config.ToolchainDigest, ConfigProfileDigest: input.Config.ProfileDigest, ConfigSnapshotDigest: input.Config.SnapshotDigest, ExpectedProviderDigest: input.Config.ExpectedProviderDigest, ExpectedObserverDigest: input.Config.ExpectedObserverDigest, BaselineDigest: input.Config.Baseline.Digest, ManifestComplete: input.Manifest.Complete, ManifestZeroChange: input.Manifest.ZeroChange, ManifestRegistryDigest: input.Manifest.RegistryDigest, ManifestToolchainDigest: input.Manifest.ToolchainDigest, ManifestProfileDigest: input.Manifest.ProfileDigest, ManifestBeforeSnapshotDigest: input.Manifest.BeforeSnapshotDigest, ManifestAfterSnapshotDigest: input.Manifest.AfterSnapshotDigest, ManifestDigest: input.Manifest.Digest, PathDigest: bridgeHash(input.InferencePath.Canonical())}
	for _, surface := range input.Registry.Surfaces {
		b.Surfaces = append(b.Surfaces, productionSurfaceBinding{SurfaceID: surface.SurfaceID.String(), CodeSymbolID: surface.CodeSymbolID.String(), SemanticOwnerID: surface.SemanticOwnerID.String(), SourceMapID: surface.Binding.SourceMapID.String(), BindingDigest: surface.Binding.BindingDigest})
	}
	for _, entry := range input.Manifest.Entries {
		b.ManifestEntries = append(b.ManifestEntries, productionManifestBinding{SurfaceID: entry.SurfaceID.String(), CodeSymbolID: entry.CodeSymbolID.String(), SemanticOwnerID: entry.SemanticOwnerID.String(), BeforeBindingDigest: entry.BeforeBindingDigest, AfterBindingDigest: entry.AfterBindingDigest, BeforeBlobDigest: entry.BeforeBlobDigest, AfterBlobDigest: entry.AfterBlobDigest})
	}
	for _, receipt := range input.Receipts {
		r := productionReceiptBinding{Schema: receipt.Schema, ReceiptID: receipt.ReceiptID.String(), SurfaceID: receipt.SurfaceID.String(), SemanticOwnerID: receipt.SemanticOwnerID.String(), CodeSymbolID: receipt.CodeSymbolID.String(), SourceMapBindingDigest: receipt.SourceMapBindingDigest, SnapshotDigest: receipt.SnapshotDigest, RegistryDigest: receipt.RegistryDigest, ToolchainDigest: receipt.ToolchainDigest, ProfileDigest: receipt.ProfileDigest, BeforeBlobDigest: receipt.BeforeBlobDigest, AfterBlobDigest: receipt.AfterBlobDigest, BeforeAuthoritySourceDigest: receipt.BeforeAuthoritySourceDigest, AfterAuthoritySourceDigest: receipt.AfterAuthoritySourceDigest, BeforeSemanticDigest: receipt.BeforeCanonicalSemanticDigest, AfterSemanticDigest: receipt.AfterCanonicalSemanticDigest, ChangeClaim: receipt.ChangeClaim, ReceiptKind: receipt.ReceiptKind, CanonicalDelta: receipt.CanonicalDelta, DeltaDigest: receipt.DeltaDigest, ClaimID: receipt.InferenceClaimID.String(), State: receipt.State}
		if receipt.AuthoritativeSource != nil {
			r.AuthoritySourceID, r.AuthoritySourcePath = receipt.AuthoritativeSource.SourceID.String(), receipt.AuthoritativeSource.Path
		}
		for _, id := range receipt.OriginPathIDs {
			r.OriginPathIDs = append(r.OriginPathIDs, id.String())
		}
		for _, ref := range receipt.EvidenceRefs {
			r.EvidenceIDs = append(r.EvidenceIDs, ref.ID.String())
			r.EvidenceDigests = append(r.EvidenceDigests, ref.Digest)
		}
		b.Receipts = append(b.Receipts, r)
	}
	if input.ExternalReceipt != nil {
		b.ExternalReceiptDigest = input.ExternalReceipt.Digest
		r := input.ExternalReceipt
		b.ExternalReceipt = productionExternalBinding{Schema: r.Schema, SnapshotDigest: r.SnapshotDigest, ProviderDigest: r.ProviderDigest, ObserverDigest: r.ObserverDigest, CPUWorkUnits: r.CPUWorkUnits, PeakMemoryBytes: r.PeakMemoryBytes, DeterministicWorkUnits: r.DeterministicWorkUnits, Digest: r.Digest}
	}
	return b
}

func bridgeDecision(status production.Status) Decision {
	switch status {
	case production.StatusPass:
		return DecisionPass
	case production.StatusFailClosed:
		return DecisionFailClosed
	default:
		return DecisionUnknown
	}
}

func bridgeBindingDigestValues(surface, code, owner, sourceMap string) string {
	var b strings.Builder
	bridgeField(&b, surface)
	bridgeField(&b, code)
	bridgeField(&b, owner)
	bridgeField(&b, sourceMap)
	return bridgeHash(b.String())
}
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
func bridgeField(b *strings.Builder, value string) {
	b.WriteString(strconv.Itoa(len(value)))
	b.WriteByte(':')
	b.WriteString(value)
	b.WriteByte('|')
}
func bridgeHash(value string) string {
	return strings.TrimPrefix(semantic.StableHashString(value), "sha256:")
}
func bridgeRawDigest(value string) string { return strings.TrimPrefix(value, "sha256:") }
func bridgeID(value string) semantic.ID   { return semantic.MustIdentity(value) }
func bridgeSourceMapID(value string) semantic.ID {
	return semantic.MustIdentity("urn:gooo:source-map:" + value)
}
func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
func registrySurface(registry production.Registry, id string) production.Surface {
	for _, s := range registry.Surfaces {
		if s.SurfaceID.String() == id {
			return s
		}
	}
	return production.Surface{}
}

func rawBeforeBlobDigest(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		for _, binding := range input.Registry {
			if binding.RegisteredSurfaceID == receipt.SurfaceID && binding.CodeSymbolID == change.CodeSymbolID {
				return change.BeforeDigest
			}
		}
	}
	return digestText("bridge-before-blob:" + receipt.SurfaceID)
}

func rawAfterBlobDigest(input Input, receipt CouplingReceipt) string {
	for _, change := range input.Changes {
		for _, binding := range input.Registry {
			if binding.RegisteredSurfaceID == receipt.SurfaceID && binding.CodeSymbolID == change.CodeSymbolID {
				return change.AfterDigest
			}
		}
	}
	return digestText("bridge-after-blob:" + receipt.SurfaceID)
}

// These tables are literal contract data. They must never call either
// evaluator to derive expected decisions, reasons, counts, resources, or
// digests. They are filled with immutable vectors for every corpus row and
// mutation before this bridge is publishable.
