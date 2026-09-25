package lsp

import (
	"encoding/json"
	"testing"
)

func assertRawReferenceResult(t *testing.T, payload []byte, id int, want string) {
	t.Helper()
	var response struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	if response.ID != id || string(response.Result) != want {
		t.Fatalf("reference response = %#v, want id=%d result=%s", response, id, want)
	}
}
func rotateReferences(references []Reference, offset int) []Reference {
	result := make([]Reference, len(references))
	for index := range references {
		result[index] = references[(index+offset)%len(references)]
	}
	return result
}
