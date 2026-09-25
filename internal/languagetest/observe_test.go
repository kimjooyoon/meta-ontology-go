package languagetest

import "testing"

const positiveSource = `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
entity PayOrderProducesPayment id "gooo://test/activity/PayOrder/output/Payment"
activity PayOrder(Order) -> Payment
`

func TestObservePassesDeclaredOutputAssertion(t *testing.T) {
	receipt := Observe(Request{Filename: "main.gooo", Source: positiveSource})
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionPass || receipt.Summary.Passed != 1 || receipt.Cases[0].Observed.Name != "Payment" {
		t.Fatalf("unexpected receipt: %+v", receipt)
	}
}

func TestObserveFailsClosedOnWrongOutput(t *testing.T) {
	source := positiveSource + "entity Refund id \"billing://entity/refund\"\nactivity RefundOrder(Order) -> Refund\n"
	source = replaceMarkerOutput(source, "Payment", "Refund")
	receipt := Observe(Request{Filename: "main.gooo", Source: source})
	if receipt.Decision != DecisionFailClosed || receipt.Reason != "LANGUAGE_TEST_ASSERTION_FAILED" || receipt.Summary.Failed != 1 {
		t.Fatalf("unexpected receipt: %+v", receipt)
	}
}

func TestObserveRejectsMissingTests(t *testing.T) {
	receipt := Observe(Request{Filename: "main.gooo", Source: `package p
namespace p
entity A id "p://a"
activity Build() -> A
`})
	if receipt.Decision != DecisionFailClosed || receipt.Reason != "LANGUAGE_TESTS_MISSING" {
		t.Fatalf("unexpected receipt: %+v", receipt)
	}
}

func replaceMarkerOutput(source, from, to string) string {
	old := "/output/" + from + "\""
	newValue := "/output/" + to + "\""
	for index := 0; index+len(old) <= len(source); index++ {
		if source[index:index+len(old)] == old {
			return source[:index] + newValue + source[index+len(old):]
		}
	}
	return source
}
