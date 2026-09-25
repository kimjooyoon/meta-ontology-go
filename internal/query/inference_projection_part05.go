package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func inferenceRecordMatches(query InferenceQuery, record semantic.InferenceRecord, kind semantic.InferenceKind, claimKind semantic.SemanticChangeKind, isClaim bool) bool {
	if query.RecordID != "" && query.RecordID != ID(record.RecordID.String()) {
		return false
	}
	if query.SubjectID != "" && query.SubjectID != ID(record.SubjectID.String()) {
		return false
	}
	if query.ObjectID != "" && query.ObjectID != ID(record.ObjectID.String()) {
		return false
	}
	if query.Kind != "" && (isClaim || query.Kind != kind) {
		return false
	}
	if query.Phase != "" && query.Phase != record.Phase.Phase {
		return false
	}
	if query.Layer != "" && query.Layer != record.Authority.Layer {
		return false
	}
	if query.Effect != "" && query.Effect != record.Authority.Effect {
		return false
	}
	if query.ClaimKind != "" && (!isClaim || query.ClaimKind != claimKind) {
		return false
	}
	if !snapshotsMatch(query.Before, record.Before) || !snapshotsMatch(query.After, record.After) {
		return false
	}
	if !controlsEmpty(query.Controls) && !controlsEqual(query.Controls, record.Controls) {
		return false
	}
	return true
}
func evidenceReferencesMatch(id ID, refs []semantic.EvidenceReference) bool {
	if id == "" {
		return true
	}
	for _, ref := range refs {
		if ID(ref.ID.String()) == id {
			return true
		}
	}
	return false
}
