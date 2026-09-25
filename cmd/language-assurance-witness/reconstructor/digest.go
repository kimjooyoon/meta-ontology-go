package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digest(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func validSHA(value string) bool {
	_, err := hex.DecodeString(value)
	return len(value) == 40 && err == nil
}

func observed(observed bool, value int) *int {
	if !observed {
		return nil
	}
	return &value
}
