package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validRecord(record semantic.InferenceRecord, evidence map[semantic.ID]semantic.InferenceEvidence, beforeDigest, afterDigest string, config EvaluationConfig) bool {
	if !validID(record.RecordID.String()) || !validID(record.SubjectID.String()) || !validID(record.ObjectID.String()) ||
		!validID(record.Rule.ID.String()) || !validToken(record.Rule.Version) || !validDigest(record.Rule.Digest) ||
		!record.Phase.Phase.Valid() || record.Phase.Ordinal == 0 || record.Before.Semantic != beforeDigest || record.After.Semantic != afterDigest ||
		!validSnapshot(record.Before) || !validSnapshot(record.After) || !record.Authority.Layer.Valid() || !record.Authority.Effect.Valid() ||
		len(record.Evidence) == 0 {
		return false
	}
	if !validControls(record.Controls, config) {
		return false
	}
	seen := make(map[semantic.ID]struct{}, len(record.Evidence))
	for _, ref := range record.Evidence {
		if !validID(ref.ID.String()) || !validDigest(ref.Digest) {
			return false
		}
		if _, duplicate := seen[ref.ID]; duplicate {
			return false
		}
		seen[ref.ID] = struct{}{}
		ev, ok := evidence[ref.ID]
		if !ok || ev.Digest != ref.Digest || ev.Before != record.Before || ev.After != record.After || !controlsEqual(ev.Controls, record.Controls) {
			return false
		}
	}
	return true
}
func validKindBinding(edge semantic.InferenceEdge) bool {
	if edge.Kind == semantic.InferenceAuthoritativeDeclaration {
		return edge.Phase.Phase == semantic.PhaseDeclaration && edge.Authority.Layer == semantic.AuthoritySource && edge.Authority.Effect == semantic.AuthorityDeclare
	}
	if edge.Kind == semantic.InferenceDeterministicDerivation {
		return edge.Phase.Phase == semantic.PhaseDerivation && edge.Authority.Layer == semantic.AuthoritySemantic && edge.Authority.Effect == semantic.AuthorityDerive
	}
	if edge.Kind == semantic.InferenceDerivedProjection {
		return edge.Phase.Phase == semantic.PhaseProjection && edge.Authority.Layer == semantic.AuthorityDerived && edge.Authority.Effect == semantic.AuthorityProject && edge.Controls.Profile.ID != ""
	}
	if edge.Kind == semantic.InferenceObservationCandidate {
		return edge.Phase.Phase == semantic.PhaseObservation && edge.Authority.Layer == semantic.AuthorityCandidate && edge.Authority.Effect == semantic.AuthorityObserve && edge.Controls.CatalogDigest != ""
	}
	if edge.Kind == semantic.InferenceAcceptedLift {
		return edge.Phase.Phase == semantic.PhaseLift && edge.Authority.Layer == semantic.AuthoritySemantic && edge.Authority.Effect == semantic.AuthorityLift && edge.Controls.PolicyDigest != ""
	}
	return edge.Phase.Phase == semantic.PhaseVerification && edge.Authority.Layer == semantic.AuthorityVerification && edge.Authority.Effect == semantic.AuthorityVerify && edge.Controls.PolicyDigest != ""
}
func validControls(controls semantic.InferenceControls, config EvaluationConfig) bool {
	if controls.CatalogDigest != "" && !validDigest(controls.CatalogDigest) || controls.PolicyDigest != "" && !validDigest(controls.PolicyDigest) {
		return false
	}
	profile := controls.Profile
	if profile.ID == "" && profile.Version == "" && profile.Digest == "" {
		return true
	}
	if profile.ID == "" || profile.Version == "" || !validDigest(profile.Digest) {
		return false
	}
	if config.Profile.ID != "" && (profile.ID != config.Profile.ID || profile.Version != config.Profile.Version || profile.Digest != config.Profile.Digest) {
		return false
	}
	return true
}
func validSnapshot(snapshot semantic.SnapshotDigests) bool {
	return (snapshot.Source != "" && validDigest(snapshot.Source)) || (snapshot.Semantic != "" && validDigest(snapshot.Semantic))
}
func controlsEqual(left, right semantic.InferenceControls) bool {
	return left.CatalogDigest == right.CatalogDigest && left.PolicyDigest == right.PolicyDigest && left.Profile == right.Profile
}
