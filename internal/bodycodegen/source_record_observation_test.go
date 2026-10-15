package bodycodegen

import (
	"context"
	"testing"
)

const sourceRecordObservationFixture = `package refunds
namespace refunds
entity Boolean id "refunds://boolean"
entity Order id "refunds://order" fields {
 field amount id "refunds://order/amount" type integer required one
 field label id "refunds://order/label" type string required one
 field approved id "refunds://order/approved" type boolean required one
}
activity Decide(Order,Boolean) -> Order computes "let result = input0; if input1 { result.amount = input0.amount; result.approved = true }; return result" assembling {
 choice "refund" field_update at "0" alternative "0 - input0.amount" intent "Record an approved refund as a negative amount."
 value_case "[{\"amount\":0,\"approved\":false,\"label\":\"seed\"},true]" -> "{\"amount\":0,\"approved\":true,\"label\":\"seed\"}"
 holdout_value_case "[{\"amount\":30,\"approved\":false,\"label\":\"반품\"},true]" -> "{\"amount\":-30,\"approved\":true,\"label\":\"반품\"}"
 holdout_value_case "[{\"amount\":40,\"approved\":false,\"label\":\"keep\"},false]" -> "{\"amount\":40,\"approved\":false,\"label\":\"keep\"}"
 observe {
  value_input "[{\"amount\":17,\"approved\":false,\"label\":\"확인\"},true]"
  value_input "[{\"amount\":61,\"approved\":false,\"label\":\"keep\"},false]"
  oracle "Expected"
  max_candidates "2"
  max_rounds "1"
  reuse_outputs "true"
  resolve_unique "true"
 }
 attempts "2"
}
activity Expected(Order,Boolean) -> Order computes "let result = input0; if input1 { result.amount = 0 - input0.amount; result.approved = true }; return result"
`

func TestSourceRecordObservationDistinguishesRefund(t *testing.T) {
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "observe-record.gooo", []byte(sourceRecordObservationFixture), "Decide")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.SelectedMask != 1 || r.Passed != 2 || r.Total != 2 || len(r.Attempts) != 1 || r.ModelCalls != 0 {
		t.Fatal("source question did not distinguish the permitted bodies", r)
	}
	if r.Holdout == nil || r.Holdout.Passed != 2 || r.Holdout.Total != 2 || r.Holdout.FieldsPassed != 6 {
		t.Fatal("separate evaluation lost", r.Holdout)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "observe-record.gooo", []byte(sourceRecordObservationFixture), result); err != nil {
		t.Fatal(err)
	}
}
