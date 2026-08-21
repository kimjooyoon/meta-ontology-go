package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func fixturePath(owner, code, surface, authority, projection, verification, sourceID, authEvidence, projectionEvidence, verificationEvidence semantic.ID, beforeSemantic, afterSemantic, beforeSource, afterSource string, config Config) semantic.InferencePathV1 {
	rule := semantic.RuleBinding{ID: fixtureID("rule"), Version: "v1", Digest: fixtureDigest("rule")}
	base := func(recordID, subject, object semantic.ID, phase semantic.InferencePhase, ordinal uint64, kind semantic.InferenceKind, evidence semantic.ID, controls semantic.InferenceControls) semantic.InferenceEdge {
		return semantic.InferenceEdge{RecordID: recordID, SubjectID: subject, ObjectID: object, Rule: rule, Phase: semantic.PhasePlacement{Phase: phase, Ordinal: ordinal}, Before: semantic.SnapshotDigests{Source: beforeSource, Semantic: beforeSemantic}, After: semantic.SnapshotDigests{Source: afterSource, Semantic: afterSemantic}, Authority: semantic.AuthorityBinding{Layer: layerFor(kind), Effect: effectFor(kind)}, Evidence: []semantic.EvidenceReference{{ID: evidence, Digest: fixtureDigest("evidence-" + evidence.String()[len("gooo://coupling/"):])}}, Controls: controls, Kind: kind, SourceRoots: []semantic.ID{sourceID}}
	}
	authControls := semantic.InferenceControls{}
	projectionControls := semantic.InferenceControls{Profile: semantic.ProfileBinding{ID: "coupling-profile", Version: "v1", Digest: config.ProfileDigest}}
	verificationControls := semantic.InferenceControls{PolicyDigest: config.ProfileDigest}
	edges := []semantic.InferenceEdge{
		base(authority, owner, code, semantic.PhaseDeclaration, 1, semantic.InferenceAuthoritativeDeclaration, authEvidence, authControls),
		base(projection, code, surface, semantic.PhaseProjection, 2, semantic.InferenceDerivedProjection, projectionEvidence, projectionControls),
		base(verification, surface, verificationEvidence, semantic.PhaseVerification, 3, semantic.InferenceIndependentVerification, verificationEvidence, verificationControls),
	}
	edges[0].SourceRoots = []semantic.ID{sourceID}
	if beforeSemantic == afterSemantic {
		edges[0].After = edges[0].Before
	}
	if beforeSource == afterSource {
		edges[0].After.Source = edges[0].Before.Source
	}
	evidence := []semantic.InferenceEvidence{
		{ID: authEvidence, Digest: fixtureDigest("evidence-auth-evidence"), Before: edges[0].Before, After: edges[0].After, Controls: authControls, SourceBacked: true},
		{ID: projectionEvidence, Digest: fixtureDigest("evidence-projection-evidence"), Before: edges[1].Before, After: edges[1].After, Controls: projectionControls},
		{ID: verificationEvidence, Digest: fixtureDigest("evidence-verification-evidence"), Before: edges[2].Before, After: edges[2].After, Controls: verificationControls, Independent: true},
	}
	return semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}
}
func layerFor(kind semantic.InferenceKind) semantic.AuthorityLayer {
	if kind == semantic.InferenceAuthoritativeDeclaration {
		return semantic.AuthoritySource
	}
	if kind == semantic.InferenceDerivedProjection {
		return semantic.AuthorityDerived
	}
	return semantic.AuthorityVerification
}
func effectFor(kind semantic.InferenceKind) semantic.AuthorityEffect {
	if kind == semantic.InferenceAuthoritativeDeclaration {
		return semantic.AuthorityDeclare
	}
	if kind == semantic.InferenceDerivedProjection {
		return semantic.AuthorityProject
	}
	return semantic.AuthorityVerify
}
