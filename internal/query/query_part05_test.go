package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestFromSemanticIRRejectsInvalidAuthoritativeGraph(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	entity, err := semantic.NewEntity("billing://entity/order", "billing", "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(entity); err != nil {
		t.Fatal(err)
	}

	if err := ir.AddFact(semantic.NewFact(activity.ID, semantic.WasGeneratedBy, entity.ID)); err != nil {

		return
	}
	if _, err := FromSemanticIR(ir); err == nil {
		t.Fatal("invalid semantic relation was projected into query graph")
	}
}
func id(raw string) ID {
	parsed, err := ParseID(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}
func assertAdd(t *testing.T, graph *Graph, fact Fact) {
	t.Helper()
	if err := graph.Add(fact); err != nil {
		t.Fatal(err)
	}
}
func pathIDs(paths []Path) [][]ID {
	result := make([][]ID, len(paths))
	for index, path := range paths {
		result[index] = append([]ID(nil), path.IDs...)
	}
	return result
}
