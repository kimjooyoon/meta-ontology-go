package bidir

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestDocumentBindingEdgesRetainEndpointProvenance(t *testing.T) {
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
	if len(document.BindingEdges) != 1 {
		t.Fatalf("binding edges = %#v", document.BindingEdges)
	}
	edge := document.BindingEdges[0]
	if edge.SourceActivitySpan.File != "binding.gooo" || edge.SourceActivitySpan.StartLine != 7 ||
		edge.SourcePortSpan.StartColumn != 14 || edge.TargetActivitySpan.StartColumn != 24 ||
		edge.TargetPortSpan.StartColumn != 33 {
		t.Fatalf("binding edge provenance = %#v", edge)
	}
}
