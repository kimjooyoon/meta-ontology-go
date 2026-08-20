//go:build detector_bridge

package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

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
