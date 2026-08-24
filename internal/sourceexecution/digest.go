package sourceexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func seal(receipt Receipt) Receipt {
	receipt.Digest = receiptDigest(receipt)
	return receipt
}

func Marshal(receipt Receipt) ([]byte, error) {
	if err := Validate(receipt); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(receipt)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
