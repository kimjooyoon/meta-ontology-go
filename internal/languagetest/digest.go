package languagetest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	payload, err := json.Marshal(receipt)
	if err != nil {
		panic(err)
	}
	return digestBytes(payload)
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
