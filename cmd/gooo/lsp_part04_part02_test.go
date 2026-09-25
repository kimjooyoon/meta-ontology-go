package main

import (
	"testing"
)

func assertLSPResponseID(t *testing.T, payload []byte, want int) {
	t.Helper()
	var response struct {
		ID int `json:"id"`
	}
	decodeLSPJSON(t, payload, &response)
	if response.ID != want {
		t.Fatalf("response ID = %d, want %d", response.ID, want)
	}
}
