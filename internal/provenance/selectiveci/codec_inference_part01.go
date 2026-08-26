package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func idsToStrings(ids []semantic.ID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}
func stringsToIDs(values []string) []semantic.ID {
	result := make([]semantic.ID, 0, len(values))
	for _, value := range values {
		result = append(result, semantic.ID(value))
	}
	return result
}
func wireSnapshotFrom(value semantic.SnapshotDigests) wireSnapshot {
	return wireSnapshot{Source: value.Source, Semantic: value.Semantic}
}
func snapshotFromWire(value wireSnapshot) semantic.SnapshotDigests {
	return semantic.SnapshotDigests{Source: value.Source, Semantic: value.Semantic}
}
func wireBindingFrom(value SnapshotBinding) wireSnapshotBinding {
	return wireSnapshotBinding{Base: wireSnapshotFrom(value.Base), Head: wireSnapshotFrom(value.Head)}
}
func bindingFromWire(value wireSnapshotBinding) SnapshotBinding {
	return SnapshotBinding{Base: snapshotFromWire(value.Base), Head: snapshotFromWire(value.Head)}
}
func wireProfileFrom(value semantic.ProfileBinding) wireProfile {
	return wireProfile{ID: value.ID, Version: value.Version, Digest: value.Digest}
}
func profileFromWire(value wireProfile) semantic.ProfileBinding {
	return semantic.ProfileBinding{ID: value.ID, Version: value.Version, Digest: value.Digest}
}
func wireRecordFrom(value semantic.InferenceRecord) wireRecord {
	evidence := make([]wireEvidenceRef, 0, len(value.Evidence))
	for _, ref := range value.Evidence {
		evidence = append(evidence, wireEvidenceRef{ID: ref.ID.String(), Digest: ref.Digest})
	}
	sort.Slice(evidence, func(i, j int) bool { return evidence[i].ID < evidence[j].ID })
	return wireRecord{
		RecordID: value.RecordID.String(), SubjectID: value.SubjectID.String(), ObjectID: value.ObjectID.String(),
		RuleID: value.Rule.ID.String(), RuleVersion: value.Rule.Version, RuleDigest: value.Rule.Digest,
		Phase: string(value.Phase.Phase), PhaseOrdinal: value.Phase.Ordinal,
		Before: wireSnapshotFrom(value.Before), After: wireSnapshotFrom(value.After),
		AuthorityLayer: string(value.Authority.Layer), AuthorityEffect: string(value.Authority.Effect),
		Evidence: evidence, CatalogDigest: value.Controls.CatalogDigest, PolicyDigest: value.Controls.PolicyDigest,
		Profile: wireProfileFrom(value.Controls.Profile),
	}
}
func recordFromWire(value wireRecord) semantic.InferenceRecord {
	evidence := make([]semantic.EvidenceReference, 0, len(value.Evidence))
	for _, ref := range value.Evidence {
		evidence = append(evidence, semantic.EvidenceReference{ID: semantic.ID(ref.ID), Digest: ref.Digest})
	}
	return semantic.InferenceRecord{
		RecordID: semantic.ID(value.RecordID), SubjectID: semantic.ID(value.SubjectID), ObjectID: semantic.ID(value.ObjectID),
		Rule:   semantic.RuleBinding{ID: semantic.ID(value.RuleID), Version: value.RuleVersion, Digest: value.RuleDigest},
		Phase:  semantic.PhasePlacement{Phase: semantic.InferencePhase(value.Phase), Ordinal: value.PhaseOrdinal},
		Before: snapshotFromWire(value.Before), After: snapshotFromWire(value.After),
		Authority: semantic.AuthorityBinding{Layer: semantic.AuthorityLayer(value.AuthorityLayer), Effect: semantic.AuthorityEffect(value.AuthorityEffect)},
		Evidence:  evidence,
		Controls:  semantic.InferenceControls{CatalogDigest: value.CatalogDigest, PolicyDigest: value.PolicyDigest, Profile: profileFromWire(value.Profile)},
	}
}
