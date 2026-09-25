package main

import (
	"testing"
)

func lspResponseCode(t *testing.T, payload []byte) int {
	t.Helper()
	var response struct {
		Error struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	decodeLSPJSON(t, payload, &response)
	return response.Error.Code
}
