package bodyrefinement

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestCounterexamplesRequireDirectExplicitInputs(t *testing.T) {
	failed := false
	round := Round{Composition: bodyexecution.Composition{Plan: bodyexecution.CompositionPlan{
		Activities: []bodyexecution.CompositionActivity{{ID: "id", Name: "Target"}}}}}
	value := bodyexecution.CompositionDelivery{ActivityID: "id", Input: json.RawMessage(`7`), Actual: json.RawMessage(`7`),
		Expected: json.RawMessage(`8`), Passed: &failed}
	check := func(delivery bodyexecution.CompositionDelivery, count int) {
		round.Runtime.Traces = []bodyexecution.CompositionTrace{{Deliveries: []bodyexecution.CompositionDelivery{delivery}}}
		if got := counterexamples(round, "Target"); len(got) != count {
			t.Fatal("incorrect feedback boundary", len(got), count)
		}
	}
	check(value, 1)
	value.ProducerID = "upstream"
	check(value, 0)
	value.ProducerID = ""
	value.Inputs = []bodyexecution.CompositionPortDelivery{{Port: "input0", Value: json.RawMessage(`7`)}, {Port: "input1", Value: json.RawMessage(`8`), ProducerID: "upstream"}}
	check(value, 0)
	value.Inputs[1].ProducerID = ""
	check(value, 1)
	value.Passed = nil
	check(value, 0)
}
