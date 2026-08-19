package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestSelectedInferenceChainRejectsMalformedGraphs(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*couplingFixture)
	}{
		{"disconnected", mutateDisconnected},
		{"fork", mutateFork},
		{"cycle", mutateCycle},
		{"wrong endpoint", mutateWrongEndpoint},
		{"omitted evidence", mutateOmittedEvidence},
		{"extra unrelated evidence", mutateExtraEvidence},
		{"duplicate evidence", mutateDuplicateEvidence},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newFixture(t, ChangeClaimDelta)
			tc.mutate(&fixture)
			result := Evaluate(fixture.input, fixture.authorityContext)
			if result.Status != StatusFailClosed || len(result.AcceptedSurfaceIDs) != 0 {
				t.Fatalf("malformed %s result = %#v", tc.name, result)
			}
		})
	}
}
func TestOriginPathIDPermutationUsesGraphOrder(t *testing.T) {
	left := newFixture(t, ChangeClaimDelta)
	right := newFixture(t, ChangeClaimDelta)
	right.input.Receipts[0].OriginPathIDs = []semantic.ID{left.authority, left.projection, left.verification}
	leftResult, rightResult := Evaluate(left.input, left.authorityContext), Evaluate(right.input, right.authorityContext)
	if leftResult.Status != StatusPass || rightResult.Status != StatusPass || leftResult.Digest != rightResult.Digest {
		t.Fatalf("path ID order changed result: left=%#v right=%#v", leftResult, rightResult)
	}
}
func mutateDisconnected(fixture *couplingFixture) {
	edge := fixture.input.InferencePath.Edges[1]
	edge.RecordID = fixtureID("disconnected-edge")
	edge.SubjectID, edge.ObjectID = fixtureID("disconnected-subject"), fixtureID("disconnected-object")
	fixture.input.InferencePath.Edges = append(fixture.input.InferencePath.Edges, edge)
	fixture.input.Receipts[0].OriginPathIDs = append(fixture.input.Receipts[0].OriginPathIDs, edge.RecordID)
}
func mutateFork(fixture *couplingFixture) {
	edge := fixture.input.InferencePath.Edges[1]
	edge.RecordID, edge.ObjectID = fixtureID("fork-edge"), fixtureID("fork-surface")
	fixture.input.InferencePath.Edges = append(fixture.input.InferencePath.Edges, edge)
	fixture.input.Receipts[0].OriginPathIDs = append(fixture.input.Receipts[0].OriginPathIDs, edge.RecordID)
}
func mutateCycle(fixture *couplingFixture) {
	edge := fixture.input.InferencePath.Edges[1]
	edge.RecordID, edge.SubjectID, edge.ObjectID = fixtureID("cycle-edge"), fixture.verification, fixture.code
	fixture.input.InferencePath.Edges = append(fixture.input.InferencePath.Edges, edge)
	fixture.input.Receipts[0].OriginPathIDs = append(fixture.input.Receipts[0].OriginPathIDs, edge.RecordID)
}
func mutateWrongEndpoint(fixture *couplingFixture) {
	fixture.input.InferencePath.Edges[2].ObjectID = fixtureID("wrong-verification-endpoint")
}
func mutateOmittedEvidence(fixture *couplingFixture) {
	fixture.input.Receipts[0].EvidenceRefs = fixture.input.Receipts[0].EvidenceRefs[:2]
}
func mutateExtraEvidence(fixture *couplingFixture) {
	edge := fixture.input.InferencePath.Edges[1]
	extra := semantic.InferenceEvidence{ID: fixtureID("unrelated-evidence"), Digest: fixtureDigest("unrelated-evidence"), Before: edge.Before, After: edge.After, Controls: edge.Controls}
	fixture.input.InferencePath.Evidence = append(fixture.input.InferencePath.Evidence, extra)
	fixture.input.Receipts[0].EvidenceRefs = append(fixture.input.Receipts[0].EvidenceRefs, semantic.EvidenceReference{ID: extra.ID, Digest: extra.Digest})
}
func mutateDuplicateEvidence(fixture *couplingFixture) {
	fixture.input.Receipts[0].EvidenceRefs = append(fixture.input.Receipts[0].EvidenceRefs, fixture.input.Receipts[0].EvidenceRefs[0])
}
