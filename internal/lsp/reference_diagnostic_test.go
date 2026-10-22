package lsp

import (
	"context"
	"testing"
)

func TestSemanticReferenceDiagnosticUsesAuthoredOccurrence(t *testing.T) {
	source := "package billing\nnamespace billing\nentity Order id \"billing://order\"\nactivity First(Order) -> Order\nactivity Second(Missing) -> Order\n"
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "fixture.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
	got := result.Diagnostics[0].Range
	if got.Start != (Position{Line: 4, Character: 16}) || got.End != (Position{Line: 4, Character: 23}) {
		t.Fatalf("diagnostic range = %#v, want second activity's Missing reference", got)
	}
}
