package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestWorkspaceSymbolsStrictParamsAndStableEmptyResults(t *testing.T) {
	server := NewServer()
	for _, raw := range []string{"", "null", "{}", `{"query":null}`, `{"query":7}`} {
		response, _, err := server.workspaceSymbolRequest(requestEnvelope{
			ID: json.RawMessage("1"), Params: json.RawMessage(raw),
		})
		if err != nil || response == nil || response.Error == nil || response.Error.Code != invalidParams {
			t.Fatalf("raw params %q response = %#v, error = %v", raw, response, err)
		}
	}
	response, _, err := server.workspaceSymbolRequest(requestEnvelope{
		ID: json.RawMessage("2"), Params: json.RawMessage(`{"query":"missing"}`),
	})
	if err != nil || response == nil || string(response.Result) != "[]" {
		t.Fatalf("empty workspace result = %#v, error = %v", response, err)
	}
}
func TestCanonicalWorkspaceSymbolsReplayAndNoMutation(t *testing.T) {
	symbols := []WorkspaceSymbol{
		{ID: "z", Name: "Same", Location: Location{URI: "file:///z.gooo", Range: testRange(0, 0, 0, 4)}},
		{ID: "a", Name: "Same", Location: Location{URI: "file:///a.gooo", Range: testRange(0, 0, 0, 4)}},
		{ID: "b", Name: "Same", Location: Location{URI: "file:///a.gooo", Range: testRange(1, 0, 1, 4)}},
	}
	original := append([]WorkspaceSymbol(nil), symbols...)
	want := canonicalWorkspaceSymbols(symbols)
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	for replay := 0; replay < 64; replay++ {
		gotJSON, err := json.Marshal(canonicalWorkspaceSymbols(rotateWorkspaceSymbols(symbols, replay)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(gotJSON, wantJSON) {
			t.Fatalf("replay %d = %s, want %s", replay, gotJSON, wantJSON)
		}
	}
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("canonical workspace symbols mutated input: %#v", symbols)
	}
}
func rotateWorkspaceSymbols(symbols []WorkspaceSymbol, offset int) []WorkspaceSymbol {
	result := make([]WorkspaceSymbol, len(symbols))
	for index := range symbols {
		result[index] = symbols[(index+offset)%len(symbols)]
	}
	return result
}
