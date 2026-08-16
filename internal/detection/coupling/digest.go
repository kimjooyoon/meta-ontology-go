package coupling

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func configCanonical(config Config) string {
	var builder strings.Builder
	field(&builder, ConfigSchemaV1)
	field(&builder, config.RegistryDigest)
	field(&builder, config.ToolchainDigest)
	field(&builder, config.ProfileDigest)
	field(&builder, config.SnapshotDigest)
	field(&builder, config.ExpectedProviderDigest)
	field(&builder, config.ExpectedObserverDigest)
	field(&builder, BaselineSchemaV1)
	field(&builder, strconv.FormatBool(config.Baseline.FullSuiteRequired))
	field(&builder, config.Baseline.Digest)
	field(&builder, strconv.FormatBool(config.ExternalReceiptRequired))
	return builder.String()
}

func baselineCanonical(baseline BaselineConfig) string {
	var builder strings.Builder
	field(&builder, BaselineSchemaV1)
	field(&builder, strconv.FormatBool(baseline.FullSuiteRequired))
	return builder.String()
}

func applicabilityCanonical(proof ApplicabilityProof) string {
	var builder strings.Builder
	field(&builder, AuthorityContextSchemaV1)
	field(&builder, proof.Schema)
	field(&builder, proof.RegistryDigest)
	field(&builder, proof.ToolchainDigest)
	field(&builder, proof.ProfileDigest)
	field(&builder, proof.SnapshotDigest)
	field(&builder, strconv.FormatBool(proof.AllowsEmpty))
	return builder.String()
}

func authorityCanonical(authority AuthorityContext) string {
	var builder strings.Builder
	field(&builder, AuthorityContextSchemaV1)
	field(&builder, authority.Schema)
	field(&builder, registryCanonical(authority.Registry))
	field(&builder, authority.Registry.Digest)
	field(&builder, authority.ToolchainDigest)
	field(&builder, authority.ProfileDigest)
	field(&builder, authority.SnapshotDigest)
	field(&builder, authority.ExpectedProviderDigest)
	field(&builder, authority.ExpectedObserverDigest)
	field(&builder, baselineCanonical(authority.Baseline))
	field(&builder, strconv.FormatBool(authority.ExternalReceiptRequired))
	if authority.Applicability != nil {
		field(&builder, applicabilityCanonical(*authority.Applicability))
		field(&builder, authority.Applicability.Digest)
	} else {
		field(&builder, "")
	}
	return builder.String()
}

func receiptCanonical(receipt CouplingReceipt) string {
	paths := append([]semantic.ID(nil), receipt.OriginPathIDs...)
	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })
	evidence := append([]semantic.EvidenceReference(nil), receipt.EvidenceRefs...)
	sort.Slice(evidence, func(i, j int) bool { return evidence[i].ID < evidence[j].ID })
	var builder strings.Builder
	field(&builder, ReceiptSchemaV1)
	field(&builder, receipt.ReceiptID.String())
	field(&builder, receipt.SurfaceID.String())
	field(&builder, receipt.SemanticOwnerID.String())
	field(&builder, receipt.CodeSymbolID.String())
	field(&builder, receipt.SourceMapBindingDigest)
	field(&builder, receipt.SnapshotDigest)
	field(&builder, receipt.RegistryDigest)
	field(&builder, receipt.ToolchainDigest)
	field(&builder, receipt.ProfileDigest)
	field(&builder, receipt.BeforeBlobDigest)
	field(&builder, receipt.AfterBlobDigest)
	field(&builder, receipt.BeforeAuthoritySourceDigest)
	field(&builder, receipt.AfterAuthoritySourceDigest)
	field(&builder, receipt.BeforeCanonicalSemanticDigest)
	field(&builder, receipt.AfterCanonicalSemanticDigest)
	field(&builder, string(receipt.ChangeClaim))
	field(&builder, string(receipt.ReceiptKind))
	field(&builder, receipt.CanonicalDelta)
	field(&builder, receipt.DeltaDigest)
	if receipt.AuthoritativeSource != nil {
		field(&builder, receipt.AuthoritativeSource.SourceID.String())
	} else {
		field(&builder, "")
	}
	for _, path := range paths {
		field(&builder, path.String())
	}
	field(&builder, receipt.InferenceClaimID.String())
	for _, ref := range evidence {
		field(&builder, ref.ID.String())
		field(&builder, ref.Digest)
	}
	field(&builder, receipt.State)
	return builder.String()
}

func externalCanonical(receipt ExternalResourceReceipt) string {
	var builder strings.Builder
	field(&builder, ResourceSchemaV1)
	field(&builder, receipt.SnapshotDigest)
	field(&builder, receipt.ProviderDigest)
	field(&builder, receipt.ObserverDigest)
	if receipt.CPUWorkUnits != nil {
		field(&builder, "cpu_work_units")
		field(&builder, strconv.FormatUint(*receipt.CPUWorkUnits, 10))
	}
	if receipt.PeakMemoryBytes != nil {
		field(&builder, "peak_memory_bytes")
		field(&builder, strconv.FormatUint(*receipt.PeakMemoryBytes, 10))
	}
	if receipt.DeterministicWorkUnits != nil {
		field(&builder, "deterministic_work_units")
		field(&builder, strconv.FormatUint(*receipt.DeterministicWorkUnits, 10))
	}
	return builder.String()
}

func resultCanonical(result Result) string {
	ids := append([]semantic.ID(nil), result.AcceptedSurfaceIDs...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	reasons := append([]Reason(nil), result.Reasons...)
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Code != reasons[j].Code {
			return reasons[i].Code < reasons[j].Code
		}
		return reasons[i].Detail < reasons[j].Detail
	})
	var builder strings.Builder
	field(&builder, ResultSchemaV1)
	field(&builder, string(result.Status))
	field(&builder, result.InputDigest)
	for _, id := range ids {
		field(&builder, id.String())
	}
	for _, reason := range reasons {
		field(&builder, string(reason.Code))
		field(&builder, reason.Detail)
	}
	writeDimension := func(dimension CountDimension) {
		field(&builder, strconv.FormatBool(dimension.Known))
		field(&builder, strconv.FormatUint(dimension.Value, 10))
	}
	writeDimension(result.Observation.ChangedSurfaces)
	writeDimension(result.Observation.Receipts)
	writeDimension(result.Observation.InferenceRecords)
	writeDimension(result.Observation.InferencePaths)
	writeDimension(result.Observation.DeterministicWork)
	writeDimension(result.Observation.ResourceWork)
	writeDimension(result.Observation.CPU)
	writeDimension(result.Observation.Memory)
	field(&builder, strconv.FormatBool(result.FullSuiteRequired))
	return builder.String()
}

func stableDigest(value string) string { return semantic.StableHashString(value) }

func inputIdentityDigest(input Input, authority AuthorityContext) string {
	receipts := append([]CouplingReceipt(nil), input.Receipts...)
	sort.Slice(receipts, func(i, j int) bool { return receiptCanonical(receipts[i]) < receiptCanonical(receipts[j]) })
	var builder strings.Builder
	field(&builder, InputSchemaV1)
	field(&builder, input.Schema)
	field(&builder, authorityCanonical(authority))
	field(&builder, configCanonical(input.Config))
	field(&builder, registryCanonical(input.Registry))
	field(&builder, input.Registry.Digest)
	field(&builder, manifestCanonical(input.Manifest))
	field(&builder, input.Manifest.Digest)
	for _, receipt := range receipts {
		field(&builder, receiptCanonical(receipt))
	}
	field(&builder, input.InferencePath.Canonical())
	if input.ExternalReceipt != nil {
		field(&builder, externalCanonical(*input.ExternalReceipt))
		field(&builder, input.ExternalReceipt.Digest)
	}
	return stableDigest(builder.String())
}
