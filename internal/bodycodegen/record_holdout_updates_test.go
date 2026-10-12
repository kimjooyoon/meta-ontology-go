package bodycodegen

import (
	"context"
	"errors"
	"testing"
)

func TestRecordHoldoutConditionalFieldUpdateKeepsSeparateCounts(t *testing.T) {
	source := []byte(`package refunds
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
 attempts "2"
}
`)
	g, _ := NewTypedPathGenerator("")
	r, err := g.GenerateSourceAssembly(context.Background(), "refunds.gooo", source, "Decide")
	if err != nil {
		t.Fatal(err)
	}
	selected := r.Report.RecordAssembly
	if selected.SelectedMask != 0 || selected.Passed != 1 || selected.Total != 1 || len(selected.Attempts) != 1 || selected.ModelCalls != 0 {
		t.Fatal("held outcome changed local selection", selected)
	}
	held := selected.Holdout
	if held == nil || held.Passed != 1 || held.Total != 2 || held.FieldsPassed != 5 || held.FieldsTotal != 6 {
		t.Fatal("whole-case and field outcomes were combined", held)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "refunds.gooo", source, r); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := g.GenerateSourceAssembly(ctx, "refunds.gooo", source, "Decide"); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled construction became an evaluated result", err)
	}
}

func TestRecordHoldoutScopeDescribesDeclaredTypedFields(t *testing.T) {
	g, _ := NewTypedPathGenerator("")
	r, err := g.GenerateSourceAssembly(context.Background(), "scope.gooo", []byte(recordHoldoutSource), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Report.CompletenessReceipt.Scope["domain_scope"]; got != "one source-owned pure record assembly with 1..16 typed inputs and declared scalar record fields" {
		t.Fatal("scope must describe the declared integer field", got)
	}
}
