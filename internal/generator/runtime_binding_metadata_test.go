package generator

import "testing"

func TestCopyIRPreservesRuntimeBindings(t *testing.T) {
	input := SemanticIR{
		Package: "billing",
		RuntimeBindings: []RuntimeBinding{{
			Schema:           "gooo.runtime-binding/v1",
			ProducerActivity: "billing://activity/produce",
			ProducerPort:     "result",
			ConsumerActivity: "billing://activity/consume",
			ConsumerPort:     "input",
			Entity:           "billing://entity/payment",
		}},
	}

	copy := copyIR(input)
	if len(copy.RuntimeBindings) != 1 {
		t.Fatalf("runtime bindings=%d, want 1", len(copy.RuntimeBindings))
	}
	input.RuntimeBindings[0].ProducerActivity = "billing://activity/changed"
	if copy.RuntimeBindings[0].ProducerActivity != "billing://activity/produce" {
		t.Fatalf("copy shares runtime binding storage: %#v", copy.RuntimeBindings)
	}
}
