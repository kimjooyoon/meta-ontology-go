package proofchoicealgebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func Seal(receipt Receipt) (Receipt, error) {
	receipt.Digest = ""
	data, err := json.Marshal(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.Digest = digestBytes(data)
	return receipt, nil
}
