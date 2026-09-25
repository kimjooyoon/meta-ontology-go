package main

import (
	"bytes"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestRunQueryEnvelopeUsesStableIDsAndPreservesCandidateIsolation(t *testing.T) {
	options := queryOptions{
		operation: "derived", root: "billing://entity/order",
		rule: string(queryengine.RuleUsedBy), layer: string(queryengine.LayerDeterministic),
		maxDepth: 1, maxDepthSet: true, limit: 10, limitSet: true,
	}
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	order, err := semantic.NewEntity("billing://entity/order", "billing", "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(order); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddCandidate(semantic.NewCandidateFact(activity.ID, semantic.Used, order.ID, "ambiguous")); err != nil {
		t.Fatal(err)
	}
	response, err := executeCLIQuery(ir, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DerivedDeterministic) != 0 || len(response.Result.DerivedCandidates) != 0 {
		t.Fatalf("candidate entered default authoritative result: %#v", response.Result)
	}

	graph, err := queryengine.FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	request := queryRequest(options)
	if _, err := graph.Execute(request); err != nil {
		t.Fatal(err)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("CLI query execution mutated the detached authority projection")
	}
	if _, ok := graph.NodeByName("", "Order"); ok {
		t.Fatal("display name resolved without its namespace")
	}
	if !reflect.DeepEqual(response.Metadata.GraphHash, beforeHash) {
		t.Fatalf("response graph hash = %q, want %q", response.Metadata.GraphHash, beforeHash)
	}
}
func TestRunQueryEnvelopeCanonicalizesInputPermutation(t *testing.T) {
	args := []string{"billing.gooo", "--exact", "--root", "billing://activity/pay-order", "--relation", "used", "--target", "billing://entity/order", "--limit", "10"}
	first := runQueryBytes(t, args, validSource)
	permuted := `package billing
namespace billing
activity PayOrder(Order) -> Order
entity Order id "billing://entity/order"
`
	second := runQueryBytes(t, args, permuted)
	if !bytes.Equal(first, second) {
		t.Fatalf("input declaration permutation changed canonical response:\n%s\n%s", first, second)
	}
}
