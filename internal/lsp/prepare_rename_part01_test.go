package lsp

import "testing"

func TestPrepareRenameForTargetUsesSemanticIdentityRange(t *testing.T) {
	id := "gooo://entity/order"
	symbols := []Symbol{{ID: id, Name: "Order", SelectionRange: testRange(0, 0, 0, 5)}}
	references := []Reference{{ID: id, Name: "Order", Range: testRange(1, 8, 1, 13)}}
	prepared, ok := prepareRenameForTarget("file:///example.gooo", id, "Order", symbols, references)
	if !ok {
		t.Fatal("prepare rename did not find semantic target")
	}
	if prepared.Range != symbols[0].SelectionRange || prepared.Placeholder != "Order" {
		t.Fatalf("prepared rename = %#v", prepared)
	}
}

func TestPrepareRenameForTargetRejectsUnknownSemanticIdentity(t *testing.T) {
	prepared, ok := prepareRenameForTarget("file:///example.gooo", "gooo://missing", "Missing", nil, nil)
	if ok {
		t.Fatalf("unknown semantic target was prepared: %#v", prepared)
	}
}
