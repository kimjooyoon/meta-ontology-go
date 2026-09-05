package syntax

import "testing"

const bindingFixture = `package billing
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity Produce(Order) -> Payment
activity Consume(Payment) -> Order

bind Produce.result -> Consume.input
`

func TestRuntimeBindingIsParsedAndFormattedWithoutLosingEndpoints(t *testing.T) {
	file, diagnostics := ParseFile("binding.gooo", bindingFixture)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	if len(file.Bindings) != 1 {
		t.Fatalf("bindings=%d, want 1", len(file.Bindings))
	}
	binding := file.Bindings[0]
	if binding.Producer.Activity.Name != "Produce" || binding.Producer.Port.Name != "result" ||
		binding.Consumer.Activity.Name != "Consume" || binding.Consumer.Port.Name != "input" {
		t.Fatalf("binding=%#v", binding)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	replayed, replayDiagnostics := ParseFile("binding.gooo", formatted)
	if replayDiagnostics.HasErrors() || replayed == nil || len(replayed.Bindings) != 1 {
		t.Fatalf("formatted replay diagnostics=%v file=%#v source=%q", replayDiagnostics, replayed, formatted)
	}
	replayedBinding := replayed.Bindings[0]
	if replayedBinding.Producer.Activity.Name != binding.Producer.Activity.Name || replayedBinding.Producer.Port.Name != binding.Producer.Port.Name ||
		replayedBinding.Consumer.Activity.Name != binding.Consumer.Activity.Name || replayedBinding.Consumer.Port.Name != binding.Consumer.Port.Name {
		t.Fatalf("formatted binding=%#v, original=%#v", replayedBinding, binding)
	}
}
