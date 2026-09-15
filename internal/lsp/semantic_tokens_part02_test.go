package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSemanticTokensPermutationReplayAndNoMutation(t *testing.T) {
	symbols := []Symbol{
		{Name: "later", Kind: SymbolFunction, SelectionRange: Range{Start: Position{Line: 1, Character: 4}, End: Position{Line: 1, Character: 9}}},
		{Name: "first", Kind: SymbolClass, SelectionRange: Range{Start: Position{Character: 2}, End: Position{Character: 7}}},
	}
	references := []Reference{
		{Name: "first", Range: Range{Start: Position{Line: 1, Character: 4}, End: Position{Line: 1, Character: 9}}},
		{Name: "other", Range: Range{Start: Position{Character: 10}, End: Position{Character: 15}}},
	}
	want, err := json.Marshal(semanticTokensForDocument(document{result: ParseResult{Symbols: symbols, References: references}}))
	if err != nil {
		t.Fatal(err)
	}
	originalSymbols := append([]Symbol(nil), symbols...)
	originalReferences := append([]Reference(nil), references...)
	for offset := 0; offset < len(symbols)*len(references); offset++ {
		rotatedSymbols := rotateSymbols(symbols, offset%len(symbols))
		rotatedReferences := rotateReferences(references, offset%len(references))
		got, marshalErr := json.Marshal(semanticTokensForDocument(document{result: ParseResult{
			Symbols: rotatedSymbols, References: rotatedReferences,
		}}))
		if marshalErr != nil || !bytes.Equal(got, want) {
			t.Fatalf("permutation %d: got %s want %s err=%v", offset, got, want, marshalErr)
		}
	}
	if !reflect.DeepEqual(symbols, originalSymbols) || !reflect.DeepEqual(references, originalReferences) {
		t.Fatal("semantic token projection mutated parser-owned slices")
	}
}
func TestSemanticTokensUseUTF16AndDeltaEncoding(t *testing.T) {
	start, err := OffsetToPosition("😀 Order", len("😀 "))
	if err != nil || start != (Position{Character: 3}) {
		t.Fatalf("UTF-16 start = %#v, err=%v", start, err)
	}
	result := semanticTokensForDocument(document{result: ParseResult{Symbols: []Symbol{
		{Name: "Order", Kind: SymbolClass, SelectionRange: Range{Start: start, End: Position{Character: 8}}},
		{Name: "Pay", Kind: SymbolFunction, SelectionRange: Range{Start: Position{Character: 9}, End: Position{Character: 12}}},
	}, References: []Reference{{Range: Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 6}}}}}})
	want := []uint32{0, 3, 5, 0, 0, 0, 6, 3, 1, 0, 1, 2, 4, 2, 0}
	if !reflect.DeepEqual(result.Data, want) {
		t.Fatalf("delta data = %v, want %v", result.Data, want)
	}
}
