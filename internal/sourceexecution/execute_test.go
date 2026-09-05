package sourceexecution

import "testing"

const fixtureSource = `package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`

func TestExecuteActivityContractReplaysExactly(t *testing.T) {
	request := Request{Filename: "billing.gooo", Source: fixtureSource, Entry: "PayOrder"}
	first, replay := Execute(request), Execute(request)
	if err := Validate(first); err != nil {
		t.Fatal(err)
	}
	if first.Digest != replay.Digest || first.Decision != "PASS" || len(first.Events) != 4 {
		t.Fatalf("first=%#v replay=%#v", first, replay)
	}
	if len(first.Entry.Inputs) != 2 || first.Entry.Output.ID != "billing://entity/payment" {
		t.Fatalf("entry=%#v", first.Entry)
	}
}

func TestExecuteFailuresAreExplicitAndSealed(t *testing.T) {
	tests := []struct {
		request Request
		code    string
	}{
		{Request{"billing.gooo", fixtureSource, "Missing"}, "SOURCE_ENTRY_UNKNOWN"},
		{Request{"broken.gooo", "package broken\nnamespace broken\nactivity", "Missing"}, "SOURCE_SYNTAX_INVALID"},
	}
	for _, test := range tests {
		receipt := Execute(test.request)
		if err := Validate(receipt); err != nil {
			t.Fatalf("%s: %v", test.code, err)
		}
		if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != test.code || len(receipt.Diagnostics) != 1 {
			t.Fatalf("receipt=%#v", receipt)
		}
	}
}

func TestExecuteRejectsRuntimeBindingsWithoutAPlan(t *testing.T) {
	request := Request{Filename: "binding.gooo", Source: fixtureSource + "\nbind PayOrder.result -> PayOrder.input\n", Entry: "PayOrder"}
	receipt := Execute(request)
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "SOURCE_RUNTIME_BINDINGS_UNSUPPORTED" || len(receipt.Events) != 0 {
		t.Fatalf("receipt=%#v", receipt)
	}
}
