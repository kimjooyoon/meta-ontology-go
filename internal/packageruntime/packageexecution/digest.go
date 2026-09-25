package packageexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(data)
}

func seal(receipt *Receipt) {
	receipt.Digest = ""
	receipt.Digest = digestValue(*receipt)
}

func Marshal(receipt Receipt) ([]byte, error) {
	if err := Validate(receipt); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
