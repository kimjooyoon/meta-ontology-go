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
	if binding.Producer.Activity.Name != "Observe" || binding.Producer.Port.Name != "result" ||
		binding.Consumer.Activity.Name != "Diagnose" || binding.Consumer.Port.Name != "input" {
		t.Fatalf("binding endpoints = %#v", binding)
	}
	if binding.Span.Filename != "binding.gooo" || binding.Span.Start.Line != 5 {
		t.Fatalf("binding provenance span = %#v", binding.Span)
	}
	if binding.Producer.Activity.Span.Start.Column != 6 || binding.Producer.Port.Span.Start.Column != 14 ||
		binding.Consumer.Activity.Span.Start.Column != 24 || binding.Consumer.Port.Span.Start.Column != 33 {
		t.Fatalf("binding endpoint spans = %#v", binding)
	}
}
