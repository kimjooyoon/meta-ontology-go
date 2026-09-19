package bidir

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestTypedPlanRetainsEndpointProvenanceThroughCanonicalModel(t *testing.T) {
	source := `package p
namespace n
entity Symptom id "urn:symptom"
entity Diagnosis id "urn:diagnosis"
activity Observe() -> Symptom
activity Diagnose(Symptom) -> Diagnosis
bind Observe.result -> Diagnose.input`
	file, diagnostics := syntax.ParseFile("binding.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatalf("binding diagnostics = %v", diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatalf("lower binding document = %v", err)
	}
	plan, err := CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Edges) != 1 {
		t.Fatalf("binding edges = %#v", plan.Edges)
	}
	edge := plan.Edges[0]
	if edge.Producer.Activity.Span.File != "binding.gooo" || edge.Producer.Activity.Span.StartLine != 7 ||
		edge.Producer.Port.Span.StartColumn != 14 || edge.Consumer.Activity.Span.StartColumn != 24 ||
		edge.Consumer.Port.Span.StartColumn != 33 {
		t.Fatalf("binding edge provenance = %#v", edge)
	}
	assertTypedPlanIRBinding(t, document, edge)
}

func assertTypedPlanIRBinding(t *testing.T, document Document, edge RuntimeBinding) {
	t.Helper()
	ir, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.RuntimeBindings) != 1 {
		t.Fatalf("IR binding count = %d", len(ir.RuntimeBindings))
	}
	binding := ir.RuntimeBindings[0]
	if string(binding.ProducerActivity) != string(edge.Producer.Activity.ID) ||
		string(binding.ConsumerActivity) != string(edge.Consumer.Activity.ID) ||
		string(binding.Entity) != string(edge.Entity) || edge.Entity == "" ||
		binding.ProducerPort != edge.Producer.Port.Name || binding.ConsumerPort != edge.Consumer.Port.Name ||
		binding.Span.File != edge.Span.File || binding.Span.Start.Offset != edge.Span.Start ||
		binding.Span.End.Offset != edge.Span.End {
		t.Fatalf("plan/IR binding disagreement: plan=%#v IR=%#v", edge, binding)
	}
}
