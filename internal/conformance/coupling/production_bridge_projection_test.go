//go:build detector_bridge

package coupling

import (
	"sort"
	"strings"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func detectorInputFromCanonical(input Input) production.Input {
	registry := productionRegistryFromCanonical(input)
	config := productionConfigFromCanonical(input, registry)
	manifest := productionManifestFromCanonical(input, registry, config)
	path := productionPathFromCanonical(input.Path, input.Receipts, input.Roots)
	return production.Input{
		Schema: production.InputSchemaV1, Config: config, Registry: registry, Manifest: manifest,
		Receipts: productionReceiptsFromCanonical(input, registry, config, path), InferencePath: path,
		ExternalReceipt: productionExternalReceiptFromCanonical(input, config), WorkspaceRoot: input.Config.Profile.ID,
	}
}

func productionRegistryFromCanonical(input Input) production.Registry {
	registry := production.Registry{Schema: production.RegistrySchemaV1}
	for _, raw := range input.Registry {
		surface := production.Surface{
			SurfaceID: bridgeID(raw.RegisteredSurfaceID), CodeSymbolID: bridgeID(raw.CodeSymbolID), SemanticOwnerID: bridgeID(raw.SemanticOwnerID),
			Binding:           production.SourceMapBinding{SourceMapID: bridgeSourceMapID(raw.SourceMapID), BindingDigest: bridgeBindingDigestValues(raw.RegisteredSurfaceID, raw.CodeSymbolID, raw.SemanticOwnerID, bridgeSourceMapID(raw.SourceMapID).String()), PackageLabel: raw.PackageLabel, FileLabel: raw.FileLabel, SourceSpan: raw.SourceSpan},
			PresentationLabel: raw.RegisteredSurfaceID,
		}
		registry.Surfaces = append(registry.Surfaces, surface)
	}
	registry.Digest = bridgeRegistryDigest(registry)
	return registry
}

func productionConfigFromCanonical(input Input, registry production.Registry) production.Config {
	baseline := production.BaselineConfig{Schema: production.BaselineSchemaV1, FullSuiteRequired: true}
	baseline.Digest = bridgeBaselineDigest(baseline)
	return production.Config{Schema: production.ConfigSchemaV1, RegistryDigest: registry.Digest,
		ToolchainDigest: bridgeRawDigest(input.Config.ToolchainDigest), ProfileDigest: bridgeRawDigest(input.Config.Profile.Digest),
		SnapshotDigest: bridgeRawDigest(input.Config.ResourceBinding.SnapshotDigest), ExpectedProviderDigest: bridgeRawDigest(input.Config.ResourceBinding.ProviderDigest),
		ExpectedObserverDigest: bridgeRawDigest(input.Config.ResourceBinding.ObserverDigest), Baseline: baseline, ExternalReceiptRequired: true}
}

func productionManifestFromCanonical(input Input, registry production.Registry, config production.Config) production.ChangeManifest {
	manifest := production.ChangeManifest{Schema: production.ManifestSchemaV1, Complete: input.Manifest.Complete, ZeroChange: input.Manifest.ZeroChange,
		RegistryDigest: registry.Digest, ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
		BeforeSnapshotDigest: bridgeRawDigest(input.Manifest.BeforeSnapshotDigest), AfterSnapshotDigest: config.SnapshotDigest}
	changes := make(map[string]CodeChange, len(input.Changes))
	for _, change := range input.Changes {
		changes[change.CodeSymbolID] = change
	}
	for _, surface := range registry.Surfaces {
		change := changes[surface.CodeSymbolID.String()]
		before, after := bridgeRawDigest(change.BeforeDigest), bridgeRawDigest(change.AfterDigest)
		if input.Manifest.ZeroChange && before == "" && after == "" {
			before, after = bridgeHash("unchanged:"+surface.CodeSymbolID.String()), bridgeHash("unchanged:"+surface.CodeSymbolID.String())
		}
		manifest.Entries = append(manifest.Entries, production.ManifestEntry{SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID,
			BeforeBindingDigest: surface.Binding.BindingDigest, AfterBindingDigest: surface.Binding.BindingDigest,
			BeforeBlobDigest: before, AfterBlobDigest: after,
			BeforeSourcePath: surface.Binding.FileLabel, AfterSourcePath: surface.Binding.FileLabel})
	}
	manifest.Digest = bridgeManifestDigest(manifest)
	return manifest
}

// The detector and independent oracle use different typed path projections.
// This function is the explicit fixture-to-producer projection; after it
// returns, all producer mutations operate on the raw production packet.
func productionPathFromCanonical(path semantic.InferencePathV1, rawReceipts []CouplingReceipt, roots []string) semantic.InferencePathV1 {
	if len(path.Edges) == 0 {
		return semantic.InferencePathV1{}
	}
	var raw CouplingReceipt
	if len(rawReceipts) != 0 {
		raw = rawReceipts[0]
	}
	out := semantic.InferencePathV1{Version: path.Version}
	for _, rawEdge := range path.Edges {
		edge := rawEdge
		edge.InferenceRecord = productionRecordFromCanonical(rawEdge.InferenceRecord)
		if rawEdge.Kind == semantic.InferenceAuthoritativeDeclaration && len(roots) != 0 {
			edge.SourceRoots = []semantic.ID{bridgeID(firstString(roots))}
		}
		switch rawEdge.Kind {
		case semantic.InferenceAuthoritativeDeclaration:
			edge.SubjectID, edge.ObjectID = bridgeID(raw.SemanticOwnerID), bridgeID(raw.CodeSymbolID)
		case semantic.InferenceDerivedProjection:
			edge.SubjectID, edge.ObjectID = bridgeID(raw.CodeSymbolID), bridgeID(raw.SurfaceID)
		case semantic.InferenceIndependentVerification:
			edge.SubjectID = bridgeID(raw.SurfaceID)
			if len(rawEdge.Evidence) != 0 {
				edge.ObjectID = rawEdge.Evidence[0].ID
			}
		}
		edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.SubjectID, edge.ObjectID
		out.Edges = append(out.Edges, edge)
	}
	for _, claim := range path.Claims {
		mapped := claim
		mapped.InferenceRecord = productionRecordFromCanonical(claim.InferenceRecord)
		mapped.CanonicalDelta = strings.TrimSpace(claim.CanonicalDelta)
		mapped.DeltaDigest = ""
		if mapped.CanonicalDelta != "" {
			mapped.DeltaDigest = bridgeHash(mapped.CanonicalDelta)
		}
		out.Claims = append(out.Claims, mapped)
	}
	for _, evidence := range path.Evidence {
		mapped := evidence
		mapped.Digest = bridgeRawDigest(evidence.Digest)
		mapped.Before = productionSnapshot(evidence.Before)
		mapped.After = productionSnapshot(evidence.After)
		mapped.Controls = productionControls(evidence.Controls)
		out.Evidence = append(out.Evidence, mapped)
	}
	return out
}

func productionRecordFromCanonical(record semantic.InferenceRecord) semantic.InferenceRecord {
	result := record
	result.Rule.Digest = bridgeRawDigest(record.Rule.Digest)
	result.Before = productionSnapshot(record.Before)
	result.After = productionSnapshot(record.After)
	result.Controls = productionControls(record.Controls)
	result.Evidence = make([]semantic.EvidenceReference, 0, len(record.Evidence))
	for _, ref := range record.Evidence {
		result.Evidence = append(result.Evidence, semantic.EvidenceReference{ID: ref.ID, Digest: bridgeRawDigest(ref.Digest)})
	}
	return result
}

func productionSnapshot(snapshot semantic.SnapshotDigests) semantic.SnapshotDigests {
	return semantic.SnapshotDigests{Source: bridgeRawDigest(snapshot.Source), Semantic: bridgeRawDigest(snapshot.Semantic)}
}

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

func productionSelectedEvidence(path semantic.InferencePathV1, ids []semantic.ID) []semantic.ID {
	byID := make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges))
	for _, edge := range path.Edges {
		byID[edge.RecordID] = edge
	}
	selected := make(map[semantic.ID]struct{})
	for _, id := range ids {
		for _, ref := range byID[id].Evidence {
			selected[ref.ID] = struct{}{}
		}
	}
	result := make([]semantic.ID, 0, len(selected))
	for id := range selected {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func productionEvidenceDigest(path semantic.InferencePathV1, id semantic.ID) string {
	for _, evidence := range path.Evidence {
		if evidence.ID == id {
			return bridgeRawDigest(evidence.Digest)
		}
	}
	return ""
}
