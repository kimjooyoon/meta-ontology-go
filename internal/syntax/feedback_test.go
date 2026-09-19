package syntax

import "testing"

func TestFeedbackSyntaxPreservesPhaseAndProvenance(t *testing.T) {
	source := "package p\nnamespace p\nfeedback Finish.result -> Start.input\n"
	file, diagnostics := ParseFile("feedback.gooo", source)
	if diagnostics.HasErrors() || file == nil || len(file.Bindings) != 1 {
		t.Fatalf("file=%#v diagnostics=%v", file, diagnostics)
	}
	binding := file.Bindings[0]
	if !binding.Feedback || binding.Span.Start.Line != 3 || binding.Producer.Activity.Name != "Finish" || binding.Consumer.Activity.Name != "Start" {
		t.Fatalf("feedback=%#v", binding)
	}
	clone := file.Clone()
	clone.Bindings[0].Feedback = false
	if !file.Bindings[0].Feedback {
		t.Fatal("clone changed original feedback phase")
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	replay, diagnostics := ParseFile("feedback.gooo", formatted)
	if diagnostics.HasErrors() || len(replay.Bindings) != 1 || !replay.Bindings[0].Feedback {
		t.Fatalf("feedback erased by format: %s; %v", formatted, diagnostics)
	}
}
