package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func claimWireToSemantic(claim claimWire) (semantic.SemanticChangeClaim, error) {
	record, err := recordWireToSemantic(claim.Record)
	if err != nil {
		return semantic.SemanticChangeClaim{}, err
	}
	return semantic.SemanticChangeClaim{InferenceRecord: record, Kind: semantic.SemanticChangeKind(claim.Kind), CanonicalDelta: claim.CanonicalDelta, DeltaDigest: claim.DeltaDigest}, nil
}
func recordWireFromSemantic(record semantic.InferenceRecord) (recordWire, error) {
	evidence := make([]evidenceRefWire, len(record.Evidence))
	for i, ref := range record.Evidence {
		evidence[i] = evidenceRefWire{ID: ref.ID.String(), Digest: ref.Digest}
	}
	return recordWire{RecordID: record.RecordID.String(), SubjectID: record.SubjectID.String(), ObjectID: record.ObjectID.String(), Rule: ruleWire{ID: record.Rule.ID.String(), Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: phaseWire{Phase: string(record.Phase.Phase), Ordinal: record.Phase.Ordinal}, Before: snapshotWire{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: snapshotWire{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: authorityWire{Layer: string(record.Authority.Layer), Effect: string(record.Authority.Effect)}, Evidence: evidence, Controls: controlsWireFromSemantic(record.Controls)}, nil
}
func recordWireToSemantic(record recordWire) (semantic.InferenceRecord, error) {
	recordID, err := semantic.ParseIdentity(record.RecordID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	subjectID, err := semantic.ParseIdentity(record.SubjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	objectID, err := semantic.ParseIdentity(record.ObjectID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	ruleID, err := semantic.ParseIdentity(record.Rule.ID)
	if err != nil {
		return semantic.InferenceRecord{}, err
	}
	evidence := make([]semantic.EvidenceReference, len(record.Evidence))
	for i, ref := range record.Evidence {
		id, parseErr := semantic.ParseIdentity(ref.ID)
		if parseErr != nil {
			return semantic.InferenceRecord{}, parseErr
		}
		evidence[i] = semantic.EvidenceReference{ID: id, Digest: ref.Digest}
	}
	return semantic.InferenceRecord{RecordID: recordID, SubjectID: subjectID, ObjectID: objectID, Rule: semantic.RuleBinding{ID: ruleID, Version: record.Rule.Version, Digest: record.Rule.Digest}, Phase: semantic.PhasePlacement{Phase: semantic.InferencePhase(record.Phase.Phase), Ordinal: record.Phase.Ordinal}, Before: semantic.SnapshotDigests{Source: record.Before.Source, Semantic: record.Before.Semantic}, After: semantic.SnapshotDigests{Source: record.After.Source, Semantic: record.After.Semantic}, Authority: semantic.AuthorityBinding{Layer: semantic.AuthorityLayer(record.Authority.Layer), Effect: semantic.AuthorityEffect(record.Authority.Effect)}, Evidence: evidence, Controls: controlsWireToSemantic(record.Controls)}, nil
}
func evidenceWireFromSemantic(evidence semantic.InferenceEvidence) (evidenceWire, error) {
	id, err := semantic.ParseIdentity(evidence.ID.String())
	if err != nil {
		return evidenceWire{}, err
	}
	return evidenceWire{ID: id.String(), Digest: evidence.Digest, Before: snapshotWire{Source: evidence.Before.Source, Semantic: evidence.Before.Semantic}, After: snapshotWire{Source: evidence.After.Source, Semantic: evidence.After.Semantic}, SourceBacked: evidence.SourceBacked, Independent: evidence.Independent, Controls: controlsWireFromSemantic(evidence.Controls)}, nil
}
func evidenceWireToSemantic(evidence evidenceWire) (semantic.InferenceEvidence, error) {
	id, err := semantic.ParseIdentity(evidence.ID)
	if err != nil {
		return semantic.InferenceEvidence{}, err
	}
	return semantic.InferenceEvidence{ID: id, Digest: evidence.Digest, Before: semantic.SnapshotDigests{Source: evidence.Before.Source, Semantic: evidence.Before.Semantic}, After: semantic.SnapshotDigests{Source: evidence.After.Source, Semantic: evidence.After.Semantic}, SourceBacked: evidence.SourceBacked, Independent: evidence.Independent, Controls: controlsWireToSemantic(evidence.Controls)}, nil
}
func controlsWireFromSemantic(controls semantic.InferenceControls) controlsWire {
	return controlsWire{CatalogDigest: controls.CatalogDigest, PolicyDigest: controls.PolicyDigest, Profile: profileWire{ID: controls.Profile.ID, Version: controls.Profile.Version, Digest: controls.Profile.Digest}}
}
func controlsWireToSemantic(controls controlsWire) semantic.InferenceControls {
	return semantic.InferenceControls{CatalogDigest: controls.CatalogDigest, PolicyDigest: controls.PolicyDigest, Profile: semantic.ProfileBinding{ID: controls.Profile.ID, Version: controls.Profile.Version, Digest: controls.Profile.Digest}}
}
