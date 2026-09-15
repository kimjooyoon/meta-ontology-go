package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func completeInferenceFixture() inferenceFixture {
	source := fixtureID("node/source-declaration")
	declared := fixtureID("node/declared-activity")
	derived := fixtureID("node/deterministic-derivation")
	projected := fixtureID("node/generated-projection")
	verified := fixtureID("node/independent-verification")
	declaration, declarationEvidence := manualInferenceEdge(
		semantic.InferenceAuthoritativeDeclaration, "01-declaration", source, declared,
	)
	derivation, derivationEvidence := manualInferenceEdge(
		semantic.InferenceDeterministicDerivation, "02-derivation", declared, derived,
	)
	projection, projectionEvidence := manualInferenceEdge(
		semantic.InferenceDerivedProjection, "03-projection", derived, projected,
	)
	verification, verificationEvidence := manualInferenceEdge(
		semantic.InferenceIndependentVerification, "04-verification", projected, verified,
	)
	edges := []semantic.InferenceEdge{declaration, derivation, projection, verification}
	evidence := []semantic.InferenceEvidence{
		declarationEvidence, derivationEvidence, projectionEvidence, verificationEvidence,
	}
	return inferenceFixture{
		path:  semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence},
		edges: edges, evidence: evidence,
	}
}
func exactEdgeSequence(t *testing.T, edges []semantic.InferenceEdge) []string {
	t.Helper()
	sequence := make([]string, 0, len(edges))
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			t.Fatalf("edge %s is invalid: %v", edge.RecordID, err)
		}
		sequence = append(sequence, edge.Canonical())
	}
	return sequence
}
func assertExactRecordSequence(t *testing.T, got, want []semantic.InferenceEdge) {
	t.Helper()
	gotSequence := exactEdgeSequence(t, got)
	wantSequence := exactEdgeSequence(t, want)
	if !reflect.DeepEqual(gotSequence, wantSequence) {
		t.Fatalf("record sequence changed:\n got: %#v\nwant: %#v", gotSequence, wantSequence)
	}
}
func reorderedFixture(fixture inferenceFixture) semantic.InferencePathV1 {
	return semantic.InferencePathV1{
		Version: semantic.InferencePathSchemaVersion,
		Edges:   []semantic.InferenceEdge{fixture.edges[2], fixture.edges[0], fixture.edges[3], fixture.edges[1]},
		Evidence: []semantic.InferenceEvidence{
			fixture.evidence[2], fixture.evidence[0], fixture.evidence[3], fixture.evidence[1],
		},
	}
}
func clonePath(path semantic.InferencePathV1) semantic.InferencePathV1 {
	clone := semantic.InferencePathV1{Version: path.Version,
		Edges:    append([]semantic.InferenceEdge(nil), path.Edges...),
		Claims:   append([]semantic.SemanticChangeClaim(nil), path.Claims...),
		Evidence: append([]semantic.InferenceEvidence(nil), path.Evidence...)}
	for index := range clone.Edges {
		clone.Edges[index].SourceRoots = append([]semantic.ID(nil), path.Edges[index].SourceRoots...)
		clone.Edges[index].Evidence = append([]semantic.EvidenceReference(nil), path.Edges[index].Evidence...)
	}
	return clone
}
