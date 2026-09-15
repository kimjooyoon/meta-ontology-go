package coupling

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strconv"
	"strings"
)

func receiptCanonical(receipt CouplingReceipt) string {
	paths := append([]semantic.ID(nil), receipt.OriginPathIDs...)
	slices.Sort(paths)
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
