package query

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func inferenceQueryDigest(value string) string {
	return semantic.StableHashString("query-inference/" + value)
}

func inferenceQueryID(value string) semantic.ID {
	return semantic.MustIdentity("inference-query://" + value)
}

func inferenceQueryEdge(
	kind semantic.InferenceKind, suffix string, subject, object semantic.ID,
) (semantic.InferenceEdge, semantic.InferenceEvidence) {
	semanticDigest := inferenceQueryDigest("semantic/" + suffix)
	controls := semantic.InferenceControls{}
	phase := semantic.PhasePlacement{Ordinal: 1}
	authority := semantic.AuthorityBinding{}
	switch kind {
	case semantic.InferenceAuthoritativeDeclaration:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDeclaration, semantic.AuthoritySource, semantic.AuthorityDeclare
	case semantic.InferenceDeterministicDerivation:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDerivation, semantic.AuthoritySemantic, semantic.AuthorityDerive
	case semantic.InferenceDerivedProjection:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseProjection, semantic.AuthorityDerived, semantic.AuthorityProject
		controls.Profile = semantic.ProfileBinding{ID: "query-inference.profile", Version: "1", Digest: inferenceQueryDigest("profile")}
	case semantic.InferenceObservationCandidate:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseObservation, semantic.AuthorityCandidate, semantic.AuthorityObserve
		controls.CatalogDigest = inferenceQueryDigest("catalog")
	case semantic.InferenceAcceptedLift:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseLift, semantic.AuthoritySemantic, semantic.AuthorityLift
		controls.PolicyDigest = inferenceQueryDigest("policy")
	case semantic.InferenceIndependentVerification:
		phase.Phase, authority.Layer, authority.Effect = semantic.PhaseVerification, semantic.AuthorityVerification, semantic.AuthorityVerify
		controls.PolicyDigest = inferenceQueryDigest("policy")
	default:
		panic("unsupported test inference kind")
	}
	evidenceID := inferenceQueryID("evidence/" + suffix)
	edge := semantic.InferenceEdge{
		InferenceRecord: semantic.InferenceRecord{
			RecordID: inferenceQueryID("record/" + suffix), SubjectID: subject, ObjectID: object,
			Rule:      semantic.RuleBinding{ID: inferenceQueryID("rule/v1"), Version: "1", Digest: inferenceQueryDigest("rule")},
			Phase:     phase,
			Before:    semantic.SnapshotDigests{Source: inferenceQueryDigest("source/" + suffix), Semantic: semanticDigest},
			After:     semantic.SnapshotDigests{Source: inferenceQueryDigest("source/" + suffix), Semantic: semanticDigest},
			Authority: authority, Controls: controls,
			Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: inferenceQueryDigest("evidence-payload/" + suffix)}},
		},
		Kind: kind,
	}
	if kind == semantic.InferenceAuthoritativeDeclaration {
		edge.SourceRoots = []semantic.ID{inferenceQueryID("source-root/" + suffix)}
	}
	if kind == semantic.InferenceAcceptedLift {
		edge.AcceptanceReceipt = evidenceID
	}
	evidence := semantic.InferenceEvidence{
		ID: evidenceID, Digest: edge.Evidence[0].Digest, Before: edge.Before, After: edge.After,
		SourceBacked: kind == semantic.InferenceAcceptedLift,
		Independent:  kind == semantic.InferenceIndependentVerification,
		Controls:     controls,
	}
	return edge, evidence
}

func inferenceQueryFixture(t *testing.T) (semantic.InferencePathV1, []semantic.InferenceEdge) {
	t.Helper()
	nodes := []semantic.ID{
		inferenceQueryID("node/0"), inferenceQueryID("node/1"), inferenceQueryID("node/2"),
		inferenceQueryID("node/3"), inferenceQueryID("node/4"), inferenceQueryID("node/5"),
		inferenceQueryID("node/6"),
	}
	kinds := []semantic.InferenceKind{
		semantic.InferenceAuthoritativeDeclaration,
		semantic.InferenceDeterministicDerivation,
		semantic.InferenceDerivedProjection,
		semantic.InferenceObservationCandidate,
		semantic.InferenceAcceptedLift,
		semantic.InferenceIndependentVerification,
	}
	path := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	edges := make([]semantic.InferenceEdge, 0, len(kinds))
	for i, kind := range kinds {
		edge, evidence := inferenceQueryEdge(kind, strings.ToLower(string(kind)), nodes[i], nodes[i+1])
		edges = append(edges, edge)
		path.Edges = append(path.Edges, edge)
		path.Evidence = append(path.Evidence, evidence)
	}
	noDeltaEdge, noDeltaEvidence := inferenceQueryEdge(
		semantic.InferenceDeterministicDerivation, "claim-no-delta", nodes[0], nodes[6],
	)
	noDeltaEdge.RecordID = inferenceQueryID("claim/no-delta")
	noDeltaEdge.Authority.Effect = semantic.AuthorityNoChange
	noDelta := semantic.SemanticChangeClaim{InferenceRecord: noDeltaEdge.InferenceRecord, Kind: semantic.NoSemanticDelta}
	path.Claims = append(path.Claims, noDelta)
	noDeltaEvidence.ID = inferenceQueryID("evidence/claim-no-delta")
	noDeltaEvidence.Digest = noDelta.Evidence[0].Digest
	path.Evidence = append(path.Evidence, noDeltaEvidence)

	deltaEdge, deltaEvidence := inferenceQueryEdge(
		semantic.InferenceDeterministicDerivation, "claim-delta", nodes[1], nodes[6],
	)
	deltaEdge.RecordID = inferenceQueryID("claim/delta")
	deltaEdge.After.Semantic = inferenceQueryDigest("semantic/claim-delta-after")
	deltaEdge.Authority.Effect = semantic.AuthorityDelta
	delta := semantic.SemanticChangeClaim{
		InferenceRecord: deltaEdge.InferenceRecord, Kind: semantic.SemanticDelta,
		CanonicalDelta: "semantic\tclaim-delta\tchanged", DeltaDigest: semantic.StableHashString("semantic\tclaim-delta\tchanged"),
	}
	path.Claims = append(path.Claims, delta)
	deltaEvidence.ID = inferenceQueryID("evidence/claim-delta")
	deltaEvidence.Digest = delta.Evidence[0].Digest
	deltaEvidence.After = delta.After
	path.Evidence = append(path.Evidence, deltaEvidence)
	if _, err := path.Normalized(); err != nil {
		t.Fatalf("fixture path invalid: %v", err)
	}
	return path, edges
}

func inferenceQueryRequest() InferenceQuery {
	return InferenceQuery{
		Schema: InferenceQuerySchema, Limit: 32, MaxDepth: 16, MaxWork: 256,
	}
}
