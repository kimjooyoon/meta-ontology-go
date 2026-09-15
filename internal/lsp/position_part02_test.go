package lsp

import (
	"testing"
)

func TestApplyChangesValidatesUTF16RangeLength(t *testing.T) {
	wrongLength := 1
	if _, err := applyChanges("😀", []TextDocumentContentChangeEvent{{
		Range: &Range{Start: Position{}, End: Position{Character: 2}}, RangeLength: &wrongLength, Text: "x",
	}}); err == nil {
		t.Fatal("accepted a range length that split the UTF-16 contract")
	}
	rightLength := 2
	if got, err := applyChanges("😀", []TextDocumentContentChangeEvent{{
		Range: &Range{Start: Position{}, End: Position{Character: 2}}, RangeLength: &rightLength, Text: "x",
	}}); err != nil || got != "x" {
		t.Fatalf("valid range change = %q, error = %v", got, err)
	}
}
