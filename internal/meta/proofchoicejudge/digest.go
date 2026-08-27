package proofchoicejudge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestReceipt(input receipt) (string, error) {
	input.Digest = ""
	data, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}
