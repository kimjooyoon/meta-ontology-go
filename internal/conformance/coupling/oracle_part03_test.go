package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestAdversarialPathClosureAndEvidencePartitions(t *testing.T) {
	base := testCorpus()[0].Input
	cases := []struct {
		name   string
		mutate func(*Input)
	}{
		{name: "disconnected", mutate: func(input *Input) { input.Path.Edges[1].ObjectID = semantic.MustIdentity("urn:gooo:term:disconnected") }},
		{name: "fork", mutate: func(input *Input) {
			edge := input.Path.Edges[1]
			edge.RecordID = semantic.MustIdentity("urn:gooo:path:fork")
			input.Path.Edges = append(input.Path.Edges, edge)
		}},
		{name: "cycle", mutate: func(input *Input) { input.Path.Edges[1].ObjectID = semantic.MustIdentity("urn:gooo:source:billing") }},
		{name: "wrong-endpoint", mutate: func(input *Input) {
			input.Path.Edges[len(input.Path.Edges)-1].ObjectID = semantic.MustIdentity("urn:gooo:receipt:wrong")
		}},
		{name: "receipt-evidence-omission", mutate: func(input *Input) { input.Receipts[0].EvidenceRefs = nil }},
		{name: "extra-unrelated-evidence", mutate: func(input *Input) {
			extra := input.Path.Evidence[0]
			extra.ID = semantic.MustIdentity("urn:gooo:evidence:unrelated")
			extra.Digest = digestText("evidence:unrelated")
			input.Path.Evidence = append(input.Path.Evidence, extra)
		}},
		{name: "duplicate-evidence", mutate: func(input *Input) { input.Path.Evidence = append(input.Path.Evidence, input.Path.Evidence[0]) }},
		{name: "candidate-becomes-authority", mutate: func(input *Input) {
			candidate := makeCouplingInput(false, true)
			*input = candidate
			input.Path.Edges[2].After.Semantic = digestText("different")
		}},
	}
	for _, test := range cases {
		input := cloneInput(base)
		test.mutate(&input)
		if got := Evaluate(input); got.Decision == DecisionPass {
			t.Errorf("%s mutation falsely passed: %+v", test.name, got)
		}
	}
	permuted := cloneInput(base)
	reverseEdges(permuted.Path.Edges)
	reverseClaims(permuted.Path.Claims)
	reverseEvidence(permuted.Path.Evidence)
	if got := Evaluate(permuted); got.Decision != DecisionPass || got.CanonicalOutputDigest != Evaluate(base).CanonicalOutputDigest {
		t.Fatal("reordered path IDs changed the result")
	}
}
func TestResourceBindingsAndMeasuredZero(t *testing.T) {
	base := testCorpus()[0].Input
	arbitrary := cloneInput(base)
	arbitrary.ResourceReceipts[0].ProviderDigest = digestText("arbitrary-provider")
	if got := Evaluate(arbitrary); got.Decision != DecisionUnknown || got.Reason != ReasonResourceUnbound {
		t.Fatalf("arbitrary provider was accepted: %+v", got)
	}
	missing := cloneInput(base)
	missing.ResourceReceipts[0].Present = false
	if got := Evaluate(missing); got.Decision != DecisionUnknown || got.Reason != ReasonResourceUnbound {
		t.Fatalf("missing resource presence was accepted: %+v", got)
	}
	zero := cloneInput(base)
	zero.ResourceReceipts[0].Value = 0
	zero.ResourceReceipts[0].BindingDigest = resourceBindingDigest(zero.ResourceReceipts[0])
	if got := Evaluate(zero); got.Decision != DecisionPass || got.Resources.CPUCoreNS != 0 {
		t.Fatalf("explicit measured zero rejected: %+v", got)
	}
}
