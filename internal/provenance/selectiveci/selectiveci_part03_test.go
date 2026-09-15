package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestEvaluateSelectiveCIClosurePartitions(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testFixture)
		status DecisionStatus
		code   string
	}{
		{name: "complete", status: Verified, code: CodeVerified},
		{name: "missing record", mutate: func(f *testFixture) { f.input.InferencePath.Edges = f.input.InferencePath.Edges[:2] }, status: Unknown, code: CodeMissing},
		{name: "duplicate path", mutate: func(f *testFixture) { f.input.Paths = append(f.input.Paths, f.input.Paths[0]) }, status: FailClosed, code: CodeDuplicate},
		{name: "ambiguity", mutate: func(f *testFixture) {
			branch := f.input.InferencePath.Edges[1]
			branch.RecordID, branch.ObjectID = testID("record/branch"), testID("branch-command")
			f.branch = branch
			f.input.InferencePath.Edges = append(f.input.InferencePath.Edges, branch)
			f.input.Paths[0].RecordIDs = append(f.input.Paths[0].RecordIDs, branch.RecordID)
			f.input.Paths[0].ExpectedKinds = append(f.input.Paths[0].ExpectedKinds, branch.Kind)
		}, status: FailClosed, code: CodeAmbiguous},
		{name: "cycle", mutate: func(f *testFixture) { f.input.InferencePath.Edges[2].ObjectID = f.obligation }, status: FailClosed, code: CodeCycle},
		{name: "wrong endpoint", mutate: func(f *testFixture) { f.input.Paths[0].ReceiptID = testID("receipt/wrong") }, status: FailClosed, code: CodeWrongEndpoint},
		{name: "stale snapshot", mutate: func(f *testFixture) { f.input.Snapshots.Head.Semantic = testDigest("stale-head") }, status: Unknown, code: CodeStaleSnapshot},
		{name: "candidate only", mutate: func(f *testFixture) {
			edge := &f.input.InferencePath.Edges[1]
			edge.Kind = semantic.InferenceObservationCandidate
			edge.Phase.Phase = semantic.PhaseObservation
			edge.Authority = semantic.AuthorityBinding{Layer: semantic.AuthorityCandidate, Effect: semantic.AuthorityObserve}
			edge.Controls.CatalogDigest = testDigest("catalog")
			edge.Before.Semantic = edge.After.Semantic
			f.input.InferencePath.Evidence[1].Before = edge.Before
			f.input.InferencePath.Evidence[1].After = edge.After
			f.input.InferencePath.Evidence[1].Controls = edge.Controls
		}, status: Unknown, code: CodeCandidate},
		{name: "receipt mismatch", mutate: func(f *testFixture) { f.input.CommandReceipts[0].Digest = testDigest("wrong-receipt") }, status: FailClosed, code: CodeReceiptMismatch},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := cloneFixture(completeFixture())
			if test.mutate != nil {
				test.mutate(&fixture)
			}
			got := Evaluate(fixture.input)
			if got.Status != test.status || got.Code != test.code {
				t.Fatalf("receipt status/code = %s/%s, want %s/%s: %#v", got.Status, got.Code, test.status, test.code, got)
			}
			if got.Status == Verified && got.Fallback != NoFallback {
				t.Fatalf("verified fallback mode = %q", got.Fallback)
			}
			if got.Status != Verified && (got.Fallback != FullSuite || got.VerifiedCommandCount != 0) {
				t.Fatalf("fallback receipt = %#v", got)
			}
		})
	}
}
func TestEvaluateReceiptCountsAndDigest(t *testing.T) {
	got := Evaluate(completeFixture().input)
	if got.Status != Verified || got.Fallback != NoFallback || got.VerifiedCommandCount != 1 || got.VerifiedObligationCount != 1 || got.VerifiedPathCount != 1 {
		t.Fatalf("complete receipt = %#v", got)
	}
	if got.Digest != got.ExpectedDigest() || got.Canonical() == "" {
		t.Fatalf("receipt digest/canonical mismatch: %#v", got)
	}
}
