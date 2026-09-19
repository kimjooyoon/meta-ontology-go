package syntax

import "testing"

func TestBindingParserPreservesNamedEndpoints(t *testing.T) {
	source := `package p
namespace n
activity Observe() -> Symptom
activity Diagnose(Symptom) -> Diagnosis
bind Observe.result -> Diagnose.input`
	file, diagnostics := ParseFile("binding.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatalf("binding diagnostics = %v", diagnostics)
	}
	if file == nil || len(file.Bindings) != 1 {
		t.Fatalf("bindings = %#v", file)
	}
	binding := file.Bindings[0]
	if binding.SourceActivity != "Observe" || binding.SourcePort != "result" ||
		binding.TargetActivity != "Diagnose" || binding.TargetPort != "input" {
		t.Fatalf("binding endpoints = %#v", binding)
	}
	if binding.Span.Filename != "binding.gooo" || binding.Span.Start.Line != 5 {
		t.Fatalf("binding provenance span = %#v", binding.Span)
	}
}
