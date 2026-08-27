package proofchoicejudge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digest(input receipt) (string, error) {
	input.Digest = ""
	data, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
