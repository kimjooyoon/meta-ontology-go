package lsp

import (
	"encoding/json"
	"testing"
)

func assertRawResult(t *testing.T, payload []byte, id int, want string) {
	t.Helper()
	var response struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	if response.ID != id || string(response.Result) != want {
		t.Fatalf("response = %#v, want id=%d result=%s", response, id, want)
	}
}
func rotateSymbols(symbols []Symbol, offset int) []Symbol {
	result := make([]Symbol, len(symbols))
	for index := range symbols {
		result[index] = symbols[(index+offset)%len(symbols)]
	}
	return result
}
func testRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{Start: Position{Line: startLine, Character: startCharacter}, End: Position{Line: endLine, Character: endCharacter}}
}
