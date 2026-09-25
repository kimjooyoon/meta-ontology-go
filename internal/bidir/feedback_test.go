package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const feedbackSource = `package feedback
namespace feedback
entity Integer id "feedback://entity/integer"
activity Start(Integer) -> Integer computes "int.add:0"
activity Finish(Integer) -> Integer computes "int.add:-1"
bind Start.result -> Finish.input
feedback Finish.result -> Start.input
`

func feedbackDocument(t *testing.T, source string) Document {
	t.Helper()
	file, diagnostics := syntax.ParseFile("feedback.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestFeedbackSurvivesModelIRAndPutWithoutBecomingCycle(t *testing.T) {
	document := feedbackDocument(t, feedbackSource)
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	written, err := Put(document, model)
	if err != nil || !DocumentEquivalent(document, written) {
		t.Fatalf("Put feedback: %v", err)
	}
	ir, err := LowerDocument(document)
	if err != nil || len(ir.RuntimeBindings) != 2 {
		t.Fatalf("IR feedback: %v %#v", err, ir.RuntimeBindings)
	}
	feedbacks := 0
	for _, binding := range ir.RuntimeBindings {
		if binding.Schema == semantic.RuntimeFeedbackSchema {
			feedbacks++
			if binding.Span.Start.Line != 7 || binding.Span.File != "feedback.gooo" || binding.Entity != "feedback://entity/integer" {
				t.Fatalf("feedback evidence=%#v", binding)
			}
		}
	}
	if feedbacks != 1 {
		t.Fatalf("feedback count=%d", feedbacks)
	}
}

func TestFeedbackInvalidTargetsFailInModelAndIR(t *testing.T) {
	cases := []string{
		strings.ReplaceAll(feedbackSource, "feedback Finish.result -> Start.input", "bind Finish.result -> Start.input"),
		feedbackSource + "feedback Start.result -> Start.input\n",
		strings.ReplaceAll(feedbackSource, "feedback Finish.result -> Start.input", "feedback Start.result -> Finish.input"),
		strings.ReplaceAll(feedbackSource, "feedback Finish.result", "feedback Missing.result"),
		strings.ReplaceAll(feedbackSource, "feedback Finish.result", "feedback Finish.other"),
	}
	for index, source := range cases {
		document := feedbackDocument(t, source)
		if _, err := Get(document); err == nil {
			t.Fatalf("model accepted invalid feedback %d", index)
		}
		if _, err := LowerDocument(document); err == nil {
			t.Fatalf("IR accepted invalid feedback %d", index)
		}
	}
}

func TestFeedbackPhaseIsSemanticRatherThanPresentation(t *testing.T) {
	model, err := Get(feedbackDocument(t, feedbackSource))
	if err != nil {
		t.Fatal(err)
	}
	changed := model.Clone()
	for index := range changed.RuntimeBindings {
		changed.RuntimeBindings[index].Feedback = false
	}
	if SemanticEquivalent(model, changed) || SemanticFingerprint(model) == SemanticFingerprint(changed) {
		t.Fatal("feedback phase was erased from semantic identity")
	}
}
