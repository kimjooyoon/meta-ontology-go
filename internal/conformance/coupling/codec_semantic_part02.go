package coupling

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func wireRecordFromSemantic(record semantic.InferenceRecord) wireRecord {
	return wireRecord{RecordID: record.RecordID.String(), SubjectID: record.SubjectID.String(), ObjectID: record.ObjectID.String(), Rule: wireRule{ID: record.Rule.ID.String(), Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: record.Phase.Phase.String(), Ordinal: record.Phase.Ordinal, Before: wireSnapshot{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: wireSnapshot{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: wireAuthority{Layer: record.Authority.Layer.String(), Effect: record.Authority.Effect.String()}, Evidence: evidenceRefsToWire(record.Evidence), Controls: wireControlsFromSemantic(record.Controls)}
}
func wireEvidenceFromSemantic(record semantic.InferenceEvidence) wireEvidence {
	return wireEvidence{ID: record.ID.String(), Digest: record.Digest, Before: wireSnapshot{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: wireSnapshot{Source: record.After.Source, Semantic: record.After.Semantic}, SourceBacked: record.SourceBacked, Independent: record.Independent, Controls: wireControlsFromSemantic(record.Controls)}
}
func wireControlsFromSemantic(value semantic.InferenceControls) wireControls {
	return wireControls{CatalogDigest: value.CatalogDigest, PolicyDigest: value.PolicyDigest, Profile: wireProfile{ID: value.Profile.ID, Version: value.Profile.Version, Digest: value.Profile.Digest}}
}
func evidenceRefsToWire(refs []semantic.EvidenceReference) []wireEvidenceRef {
	out := make([]wireEvidenceRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, wireEvidenceRef{ID: ref.ID.String(), Digest: ref.Digest})
	}
	return out
}
func idsToStrings(ids []semantic.ID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
func semanticEdgeFromWire(raw wireEdge) (semantic.InferenceEdge, error) {
	record, err := semanticRecordFromWire(wireRecord{RecordID: raw.RecordID, SubjectID: raw.SubjectID, ObjectID: raw.ObjectID, Rule: raw.Rule, Phase: raw.Phase, Ordinal: raw.Ordinal, Before: raw.Before, After: raw.After, Authority: raw.Authority, Evidence: raw.Evidence, Controls: raw.Controls})
	if err != nil {
		return semantic.InferenceEdge{}, err
	}
	kind := semantic.InferenceKind(raw.Kind)
	if !kind.Valid() {
		return semantic.InferenceEdge{}, fmt.Errorf("unknown inference kind %q", raw.Kind)
	}
	roots := make([]semantic.ID, 0, len(raw.SourceRoots))
	for _, root := range raw.SourceRoots {
		id, err := semantic.ParseIdentity(root)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
		roots = append(roots, id)
	}
	var receipt semantic.ID
	if raw.AcceptanceReceipt != "" {
		receipt, err = semantic.ParseIdentity(raw.AcceptanceReceipt)
		if err != nil {
			return semantic.InferenceEdge{}, err
		}
	}
	return semantic.InferenceEdge{InferenceRecord: record, Kind: kind, SourceRoots: roots, AcceptanceReceipt: receipt}, nil
}
func semanticClaimFromWire(raw wireClaim) (semantic.SemanticChangeClaim, error) {
	record, err := semanticRecordFromWire(wireRecord{RecordID: raw.RecordID, SubjectID: raw.SubjectID, ObjectID: raw.ObjectID, Rule: raw.Rule, Phase: raw.Phase, Ordinal: raw.Ordinal, Before: raw.Before, After: raw.After, Authority: raw.Authority, Evidence: raw.Evidence, Controls: raw.Controls})
	if err != nil {
		return semantic.SemanticChangeClaim{}, err
	}
	return semantic.SemanticChangeClaim{InferenceRecord: record, Kind: semantic.SemanticChangeKind(raw.Kind), CanonicalDelta: raw.CanonicalDelta, DeltaDigest: raw.DeltaDigest}, nil
}
func semanticEvidenceFromWire(raw wireEvidence) (semantic.InferenceEvidence, error) {
	id, err := semantic.ParseIdentity(raw.ID)
	if err != nil {
		return semantic.InferenceEvidence{}, err
	}
	return semantic.InferenceEvidence{ID: id, Digest: raw.Digest, Before: semantic.SnapshotDigests{Source: raw.Before.Source, Semantic: raw.Before.Semantic}, After: semantic.SnapshotDigests{Source: raw.After.Source, Semantic: raw.After.Semantic}, SourceBacked: raw.SourceBacked, Independent: raw.Independent, Controls: semanticControlsFromWire(raw.Controls)}, nil
}
