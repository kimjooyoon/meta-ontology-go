package lsp

import "testing"

func TestRenameEditsPreserveSemanticIdentityAndUseClientEdits(t *testing.T) {
	const targetID = "gooo://example/order"
	symbols := []Symbol{{
		ID: targetID, Name: "Order", SelectionRange: Range{Start: Position{Line: 0, Character: 6}, End: Position{Line: 0, Character: 11}},
	}}
	references := []Reference{
		{ID: targetID, Name: "Order", Range: Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 7}}},
		{ID: targetID, Name: "Order", Range: Range{Start: Position{Line: 2, Character: 2}, End: Position{Line: 2, Character: 7}}},
	}
	edits := renameEditsForTarget("file:///example.gooo", targetID, "Order", symbols, references, "Invoice")
	if len(edits) != 3 {
		t.Fatalf("rename edits = %#v, want declaration and two references", edits)
	}
	for _, edit := range edits {
		if edit.NewText != "Invoice" {
			t.Fatalf("rename edit text = %#v", edit)
		}
	}
}

func TestValidRenameIdentifierRejectsKeywordsAndInvalidStarts(t *testing.T) {
	for _, value := range []string{"entity", "1Order", "Order-2", ""} {
		if validRenameIdentifier(value) {
			t.Fatalf("invalid rename identifier %q was accepted", value)
		}
	}
	for _, value := range []string{"Invoice", "_Invoice2", "주문서"} {
		if !validRenameIdentifier(value) {
			t.Fatalf("valid rename identifier %q was rejected", value)
		}
	}
}
