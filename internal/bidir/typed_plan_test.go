package bidir

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const typedPlanSource = `package p
namespace p
entity Integer id "p://entity/integer"
activity Produce(Integer) -> Integer
activity Consume(Integer) -> Integer
activity Finish(Integer) -> Integer
`

const typedPlanChain = `bind Produce.result -> Consume.input
bind Consume.result -> Finish.input
`

func typedPlanDocument(t *testing.T, source string) Document {
	t.Helper()
	file, diagnostics := syntax.ParseFile("typed-plan.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatalf("parse: %v", diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestTypedPlanUsesCanonicalIdentitiesAndStableOrder(t *testing.T) {
	document := typedPlanDocument(t, typedPlanSource+typedPlanChain)
	plan, err := CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	want := []ID{"p://activity/produce", "p://activity/consume", "p://activity/finish"}
	if !reflect.DeepEqual(plan.Activities, want) || len(plan.Edges) != 2 {
		t.Fatalf("plan=%#v, want order %v and 2 explicit bindings", plan, want)
	}
	reversed := "bind Consume.result -> Finish.input\nbind Produce.result -> Consume.input\n"
	reordered, err := CompileTypedPlan(typedPlanDocument(t, typedPlanSource+reversed))
	if err != nil || !reflect.DeepEqual(plan.Activities, reordered.Activities) {
		t.Fatalf("reordered plan=%#v error=%v", reordered, err)
	}
	if len(reordered.Edges) != len(plan.Edges) {
		t.Fatalf("reordered edge count=%d", len(reordered.Edges))
	}
	for index, edge := range plan.Edges {
		if runtimeBindingKey(edge) != runtimeBindingKey(reordered.Edges[index]) ||
			edge.Entity != ID("p://entity/integer") || reordered.Edges[index].Entity != edge.Entity {
			t.Fatalf("canonical edge %d changed: %#v %#v", index, edge, reordered.Edges[index])
		}
	}
}

func TestTypedPlanRejectsInvalidBindingDocuments(t *testing.T) {
	cases := []struct {
		name   string
		suffix string
		cause  error
	}{
		{"missing-edges", "", nil},
		{"unknown-activity", "bind Missing.result -> Consume.input", nil},
		{"unknown-port", "bind Produce.other -> Consume.input", semantic.ErrRuntimeBindingPort},
		{"duplicate-edge", typedPlanChain + typedPlanChain, semantic.ErrRuntimeBindingDuplicate},
		{"cycle", "bind Produce.result -> Consume.input\nbind Consume.result -> Produce.input", semantic.ErrRuntimeBindingCycle},
		{"input-conflict", "bind Produce.result -> Finish.input\nbind Consume.result -> Finish.input", semantic.ErrRuntimeBindingInputConflict},
		{"type-mismatch", "entity Text id \"p://entity/text\"\nactivity Other(Text) -> Text\nbind Produce.result -> Other.input", semantic.ErrRuntimeBindingTypeMismatch},
		{"repeated-input", "activity Repeated(Integer, Integer) -> Integer\nbind Produce.result -> Repeated.input", semantic.ErrRuntimeBindingPort},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := CompileTypedPlan(typedPlanDocument(t, typedPlanSource+tc.suffix))
			if err == nil || !reflect.DeepEqual(plan, TypedPlan{}) {
				t.Fatalf("invalid document produced plan=%#v error=%v", plan, err)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("error=%v, want cause %v", err, tc.cause)
			}
		})
	}
}

func TestTypedPlanFailureDoesNotBlockIndependentDocument(t *testing.T) {
	valid := typedPlanDocument(t, typedPlanSource+typedPlanChain)
	before, err := CompileTypedPlan(valid)
	if err != nil {
		t.Fatal(err)
	}
	invalid := typedPlanDocument(t, typedPlanSource+"bind Missing.result -> Consume.input")
	if _, err := CompileTypedPlan(invalid); err == nil {
		t.Fatal("invalid document unexpectedly accepted")
	}
	after, err := CompileTypedPlan(valid)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("independent compilation affected by failure: before=%#v after=%#v error=%v", before, after, err)
	}
}
