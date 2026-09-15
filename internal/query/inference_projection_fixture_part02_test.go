package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

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
