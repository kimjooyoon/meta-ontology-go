package generator

import (
	"strings"
	"testing"
)

func TestRenderRuntimeBindingSupportEmitsTypedExecutionBoundary(t *testing.T) {
	ir := SemanticIR{
		Imports: []Import{{Path: "reflect", Name: "goooRuntimeReflect"}},
		Activities: []Activity{
			{ID: "billing://activity/produce", GoName: "Produce"},
			{ID: "billing://activity/consume", GoName: "Consume"},
		},
		RuntimeBindings: []RuntimeBinding{{
			Schema:           "gooo.runtime-binding/v1",
			ProducerActivity: "billing://activity/produce",
			ProducerPort:     "result",
			ConsumerActivity: "billing://activity/consume",
			ConsumerPort:     "input",
			Entity:           "billing://entity/payment",
		}},
	}

	rendered := renderRuntimeBindingSupport(ir)
	for _, want := range []string{
		"type GoooRuntimeBinding struct",
		"func ExecuteGoooRuntimeBindings",
		"RUNTIME_INPUT_TYPE_MISMATCH",
		"goooRuntimeReflect.ValueOf",
		"billing://activity/produce",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("runtime support missing %q:\n%s", want, rendered)
		}
	}
}

func TestRuntimeActivityOrderIsDeterministicTopologicalOrder(t *testing.T) {
	ir := SemanticIR{
		Activities: []Activity{
			{ID: "z"}, {ID: "a"}, {ID: "m"},
		},
		RuntimeBindings: []RuntimeBinding{
			{ProducerActivity: "z", ConsumerActivity: "m"},
			{ProducerActivity: "a", ConsumerActivity: "z"},
		},
	}
	got := runtimeActivityOrder(ir)
	want := []string{"a", "z", "m"}
	if len(got) != len(want) {
		t.Fatalf("order=%v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("order=%v, want %v", got, want)
		}
	}
}
