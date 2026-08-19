package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func productionControls(controls semantic.InferenceControls) semantic.InferenceControls {
	result := controls
	result.CatalogDigest = bridgeRawDigest(controls.CatalogDigest)
	result.PolicyDigest = bridgeRawDigest(controls.PolicyDigest)
	result.Profile.Digest = bridgeRawDigest(controls.Profile.Digest)
	return result
}
func productionReceiptsFromCanonical(input Input, registry production.Registry, config production.Config, path semantic.InferencePathV1) []production.CouplingReceipt {
	result := make([]production.CouplingReceipt, 0, len(input.Receipts))
	for _, raw := range input.Receipts {
		surface := registrySurface(registry, raw.SurfaceID)
		pathIDs := productionSelectedPathIDs(path)
		evidenceIDs := productionSelectedEvidence(path, pathIDs)
		evidenceRefs := make([]semantic.EvidenceReference, 0, len(evidenceIDs))
		for _, id := range evidenceIDs {
			evidenceRefs = append(evidenceRefs, semantic.EvidenceReference{ID: id, Digest: productionEvidenceDigest(path, id)})
		}
		var source *production.AuthoritySource
		if raw.AuthoritativeSourceRef != "" {
			source = &production.AuthoritySource{SourceID: bridgeID(firstString(input.Roots)), Path: raw.AuthoritativeSourceRef}
		}
		canonicalDelta := strings.TrimSpace(raw.SemanticDelta)
		deltaDigest := ""
		if canonicalDelta != "" {
			deltaDigest = bridgeHash(canonicalDelta)
		}
		result = append(result, production.CouplingReceipt{Schema: production.ReceiptSchemaV1, ReceiptID: bridgeID(raw.ReceiptID), SurfaceID: bridgeID(raw.SurfaceID), SemanticOwnerID: bridgeID(raw.SemanticOwnerID), CodeSymbolID: bridgeID(raw.CodeSymbolID),
			SourceMapBindingDigest: surface.Binding.BindingDigest, SnapshotDigest: bridgeRawDigest(raw.SnapshotDigest), RegistryDigest: config.RegistryDigest, ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
			BeforeBlobDigest: bridgeRawDigest(rawBeforeBlobDigest(input, raw)), AfterBlobDigest: bridgeRawDigest(rawAfterBlobDigest(input, raw)), BeforeAuthoritySourceDigest: bridgeRawDigest(raw.AuthoritySourceBeforeDigest), AfterAuthoritySourceDigest: bridgeRawDigest(raw.AuthoritySourceAfterDigest),
			BeforeCanonicalSemanticDigest: bridgeRawDigest(raw.BeforeIRDigest), AfterCanonicalSemanticDigest: bridgeRawDigest(raw.AfterIRDigest), ChangeClaim: production.ChangeClaim(raw.ChangeClaim), ReceiptKind: semantic.SemanticChangeKind(raw.ReceiptKind),
			CanonicalDelta: canonicalDelta, DeltaDigest: deltaDigest, AuthoritativeSource: source, OriginPathIDs: pathIDs, InferenceClaimID: bridgeID(raw.ClaimRecordID), EvidenceRefs: evidenceRefs, State: raw.State})
	}
	return result
}
func productionExternalReceiptFromCanonical(input Input, config production.Config) *production.ExternalResourceReceipt {
	if len(input.ResourceReceipts) == 0 {
		return nil
	}
	external := &production.ExternalResourceReceipt{Schema: production.ResourceSchemaV1, SnapshotDigest: config.SnapshotDigest, ProviderDigest: bridgeRawDigest(input.ResourceReceipts[0].ProviderDigest), ObserverDigest: bridgeRawDigest(input.ResourceReceipts[0].ObserverDigest)}
	for _, raw := range input.ResourceReceipts {
		value := raw.Value
		switch raw.Metric {
		case "cpu-core-ns":
			external.CPUWorkUnits = &value
		case "peak-memory-bytes":
			external.PeakMemoryBytes = &value
		case "work-units":
			external.DeterministicWorkUnits = &value
		}
	}
	external.Digest = bridgeExternalDigest(*external)
	return external
}
func productionSelectedPathIDs(path semantic.InferencePathV1) []semantic.ID {
	ids := make([]semantic.ID, 0, 3)
	for _, edge := range path.Edges {
		switch edge.Kind {
		case semantic.InferenceAuthoritativeDeclaration, semantic.InferenceDerivedProjection, semantic.InferenceIndependentVerification:
			ids = append(ids, edge.RecordID)
		}
	}
	return ids
}
