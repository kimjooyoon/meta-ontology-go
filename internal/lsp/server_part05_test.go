package lsp

import (
	"encoding/json"
	"testing"
)

func assertResultID(t *testing.T, payload []byte, want int) {
	t.Helper()
	var message struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &message)
	if message.ID != want || string(message.Result) != "null" {
		t.Fatalf("response = %#v", message)
	}
}
