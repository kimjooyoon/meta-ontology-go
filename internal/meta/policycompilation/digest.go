package policycompilation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func DigestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func SemanticDigest(irHash string) string {
	return "sha256:" + irHash
}

func ValidDigest(value string) bool {
	return digestPattern.MatchString(value)
}

func canonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func digestJSON(value any) (string, error) {
	encoded, err := canonicalJSON(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(encoded), nil
}

func receiptDigest(receipt Receipt) (string, error) {
	receipt.ReceiptDigest = ""
	return digestJSON(receipt)
}
