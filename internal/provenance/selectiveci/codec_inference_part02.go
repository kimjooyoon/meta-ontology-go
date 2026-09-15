package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func wirePathFrom(value Path) wirePath {
	kinds := make([]string, 0, len(value.ExpectedKinds))
	for _, kind := range value.ExpectedKinds {
		kinds = append(kinds, string(kind))
	}
	return wirePath{PathID: value.PathID.String(), RootID: value.RootID.String(), ObligationID: value.ObligationID.String(), CommandID: value.CommandID.String(), ReceiptID: value.ReceiptID.String(), RecordIDs: idsToStrings(value.RecordIDs), ExpectedKinds: kinds}
}
func pathFromWire(value wirePath) Path {
	kinds := make([]semantic.InferenceKind, 0, len(value.ExpectedKinds))
	for _, kind := range value.ExpectedKinds {
		kinds = append(kinds, semantic.InferenceKind(kind))
	}
	return Path{PathID: semantic.ID(value.PathID), RootID: semantic.ID(value.RootID), ObligationID: semantic.ID(value.ObligationID), CommandID: semantic.ID(value.CommandID), ReceiptID: semantic.ID(value.ReceiptID), RecordIDs: stringsToIDs(value.RecordIDs), ExpectedKinds: kinds}
}
func wireCommandReceiptFrom(value CommandReceipt) wireCommandReceipt {
	return wireCommandReceipt{CommandID: value.CommandID.String(), ReceiptID: value.ReceiptID.String(), Status: string(value.Status), ProviderReceiptDigest: value.ProviderReceiptDigest, PhaseReceiptDigest: value.PhaseReceiptDigest, ResourceReceiptDigest: value.ResourceReceiptDigest, RegistryDigest: value.RegistryDigest, PlanDigest: value.PlanDigest, Digest: value.Digest}
}
func commandReceiptFromWire(value wireCommandReceipt) CommandReceipt {
	return CommandReceipt{CommandID: semantic.ID(value.CommandID), ReceiptID: semantic.ID(value.ReceiptID), Status: ReceiptStatus(value.Status), ProviderReceiptDigest: value.ProviderReceiptDigest, PhaseReceiptDigest: value.PhaseReceiptDigest, ResourceReceiptDigest: value.ResourceReceiptDigest, RegistryDigest: value.RegistryDigest, PlanDigest: value.PlanDigest, Digest: value.Digest}
}
func wirePathSetFrom(value semantic.InferencePathV1) wireInferencePath {
	edges := make([]wireEdge, 0, len(value.Edges))
	for _, edge := range value.Edges {
		edges = append(edges, wireEdge{Record: wireRecordFrom(edge.InferenceRecord), Kind: string(edge.Kind), SourceRoots: idsToStrings(edge.SourceRoots), AcceptanceReceipt: edge.AcceptanceReceipt.String()})
	}
	claims := make([]wireClaim, 0, len(value.Claims))
	for _, claim := range value.Claims {
		claims = append(claims, wireClaim{Record: wireRecordFrom(claim.InferenceRecord), Kind: string(claim.Kind), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest})
	}
	evidence := make([]wireInferenceEvidence, 0, len(value.Evidence))
	for _, record := range value.Evidence {
		evidence = append(evidence, wireInferenceEvidence{ID: record.ID.String(), Digest: record.Digest, Before: wireSnapshotFrom(record.Before), After: wireSnapshotFrom(record.After), SourceBacked: record.SourceBacked, Independent: record.Independent, CatalogDigest: record.Controls.CatalogDigest, PolicyDigest: record.Controls.PolicyDigest, Profile: wireProfileFrom(record.Controls.Profile)})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].Record.RecordID < edges[j].Record.RecordID })
	sort.Slice(claims, func(i, j int) bool { return claims[i].Record.RecordID < claims[j].Record.RecordID })
	sort.Slice(evidence, func(i, j int) bool { return evidence[i].ID < evidence[j].ID })
	return wireInferencePath{Version: value.Version, Edges: edges, Claims: claims, Evidence: evidence}
}
func pathSetFromWire(value wireInferencePath) semantic.InferencePathV1 {
	edges := make([]semantic.InferenceEdge, 0, len(value.Edges))
	for _, edge := range value.Edges {
		edges = append(edges, semantic.InferenceEdge{InferenceRecord: recordFromWire(edge.Record), Kind: semantic.InferenceKind(edge.Kind), SourceRoots: stringsToIDs(edge.SourceRoots), AcceptanceReceipt: semantic.ID(edge.AcceptanceReceipt)})
	}
	claims := make([]semantic.SemanticChangeClaim, 0, len(value.Claims))
	for _, claim := range value.Claims {
		claims = append(claims, semantic.SemanticChangeClaim{InferenceRecord: recordFromWire(claim.Record), Kind: semantic.SemanticChangeKind(claim.Kind), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest})
	}
	evidence := make([]semantic.InferenceEvidence, 0, len(value.Evidence))
	for _, record := range value.Evidence {
		evidence = append(evidence, semantic.InferenceEvidence{ID: semantic.ID(record.ID), Digest: record.Digest, Before: snapshotFromWire(record.Before), After: snapshotFromWire(record.After), SourceBacked: record.SourceBacked, Independent: record.Independent, Controls: semantic.InferenceControls{CatalogDigest: record.CatalogDigest, PolicyDigest: record.PolicyDigest, Profile: profileFromWire(record.Profile)}})
	}
	return semantic.InferencePathV1{Version: value.Version, Edges: edges, Claims: claims, Evidence: evidence}
}
