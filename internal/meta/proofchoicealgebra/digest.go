package proofchoicealgebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func digestSource(source []byte) string {
	sum := sha256.Sum256(source)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestReceipt(receipt Receipt) (string, error) {
	receipt.Digest = ""
	data, err := json.Marshal(receipt)
	if err != nil {
		return "", fmt.Errorf("marshal receipt for digest: %w", err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
