package semantic

import "testing"

func TestRuntimeBindingSemanticHashIgnoresSourceOrderAndSpans(t *testing.T) {
	base := runtimeBindingFixture(t)
	base.RuntimeBindings[0].Span.Start.Offset = 10
	base.RuntimeBindings[0].Span.End.Offset = 20
	base.RuntimeBindings[1].Span.Start.Offset = 30
	base.RuntimeBindings[1].Span.End.Offset = 40
	reordered := base
	reordered.RuntimeBindings = []RuntimeBinding{base.RuntimeBindings[1], base.RuntimeBindings[0]}
	if base.SemanticCanonical() != reordered.SemanticCanonical() || base.StableHash() != reordered.StableHash() {
		t.Fatalf("semantic binding order changed: base=%q reordered=%q", base.SemanticCanonical(), reordered.SemanticCanonical())
	}
	if base.Canonical() == reordered.Canonical() {
		t.Fatal("full canonical form discarded binding source order or spans")
	}
}

func TestRuntimeBindingValidationFailsClosedForUnsupportedEdges(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*IR)
		want   error
	}{
		{name: "unknown endpoint", mutate: func(ir *IR) {
			ir.RuntimeBindings[0].ProducerActivity = mustIdentity("billing://activity/unknown")
		}, want: ErrRuntimeBindingUnknownNode},
		{name: "type mismatch", mutate: func(ir *IR) {
			ir.RuntimeBindings[0].ConsumerActivity = mustIdentity("billing://activity/other")
		}, want: ErrRuntimeBindingTypeMismatch},
		{name: "duplicate incoming", mutate: func(ir *IR) {
			ir.RuntimeBindings = append(ir.RuntimeBindings, RuntimeBinding{
				Schema: RuntimeBindingSchema, ProducerActivity: mustIdentity("billing://activity/other"), ProducerPort: RuntimeOutputPort,
				ConsumerActivity: mustIdentity("billing://activity/consume"), ConsumerPort: RuntimeInputPort,
			})
		}, want: ErrRuntimeBindingInputConflict},
		{name: "cycle", mutate: func(ir *IR) {
			ir.RuntimeBindings = []RuntimeBinding{
				{Schema: RuntimeBindingSchema, ProducerActivity: mustIdentity("billing://activity/produce"), ProducerPort: RuntimeOutputPort, ConsumerActivity: mustIdentity("billing://activity/consume"), ConsumerPort: RuntimeInputPort},
				{Schema: RuntimeBindingSchema, ProducerActivity: mustIdentity("billing://activity/consume"), ProducerPort: RuntimeOutputPort, ConsumerActivity: mustIdentity("billing://activity/produce"), ConsumerPort: RuntimeInputPort},
			}
		}, want: ErrRuntimeBindingCycle},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := runtimeBindingFixture(t)
			test.mutate(&fixture)
			if err := fixture.Validate(); err == nil || !errorsIs(err, test.want) {
				t.Fatalf("error=%v, want %v", err, test.want)
			}
		})
	}
}

func runtimeBindingFixture(t *testing.T) IR {
	t.Helper()
	namespace, err := ParseNamespace("billing")
	if err != nil {
		t.Fatal(err)
	}
	ids := func(raw string) ID { return mustIdentity(raw) }
	ir := NewIR("billing", namespace)
	for _, node := range []Node{
		{Kind: Entity, ID: ids("billing://entity/order"), Namespace: namespace, Name: "Order"},
		{Kind: Entity, ID: ids("billing://entity/payment"), Namespace: namespace, Name: "Payment"},
		{Kind: Activity, ID: ids("billing://activity/produce"), Namespace: namespace, Name: "Produce"},
		{Kind: Activity, ID: ids("billing://activity/consume"), Namespace: namespace, Name: "Consume"},
		{Kind: Activity, ID: ids("billing://activity/other"), Namespace: namespace, Name: "Other"},
	} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range []Fact{
		NewFact(ids("billing://activity/produce"), Used, ids("billing://entity/order")),
		NewFact(ids("billing://entity/payment"), WasGeneratedBy, ids("billing://activity/produce")),
		NewFact(ids("billing://activity/consume"), Used, ids("billing://entity/payment")),
		NewFact(ids("billing://entity/order"), WasGeneratedBy, ids("billing://activity/consume")),
		NewFact(ids("billing://activity/other"), Used, ids("billing://entity/order")),
		NewFact(ids("billing://entity/payment"), WasGeneratedBy, ids("billing://activity/other")),
	} {
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	ir.RuntimeBindings = []RuntimeBinding{
		{Schema: RuntimeBindingSchema, ProducerActivity: ids("billing://activity/produce"), ProducerPort: RuntimeOutputPort, ConsumerActivity: ids("billing://activity/consume"), ConsumerPort: RuntimeInputPort},
		{Schema: RuntimeBindingSchema, ProducerActivity: ids("billing://activity/consume"), ProducerPort: RuntimeOutputPort, ConsumerActivity: ids("billing://activity/other"), ConsumerPort: RuntimeInputPort},
	}
	return ir
}

func mustIdentity(raw string) ID {
	id, _ := ParseIdentity(raw)
	return id
}

func errorsIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		unwrapped, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = unwrapped.Unwrap()
	}
	return false
}
