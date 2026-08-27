package model

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func DigestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(raw)
}

func ValidDigest(value string) bool { return digestPattern.MatchString(value) }
func ValidHead(value string) bool   { return headPattern.MatchString(value) }

func SealReceipt(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = DigestJSON(receipt)
	return receipt
}

func VerifyReceiptDigest(receipt Receipt) bool {
	digest := receipt.Digest
	receipt.Digest = ""
	return ValidDigest(digest) && digest == DigestJSON(receipt)
}

func SealVerdict(verdict Verdict) Verdict {
	verdict.Digest = ""
	verdict.Digest = DigestJSON(verdict)
	return verdict
}

func VerifyVerdictDigest(verdict Verdict) bool {
	digest := verdict.Digest
	verdict.Digest = ""
	return ValidDigest(digest) && digest == DigestJSON(verdict)
}

func DecodeStrict[T any](raw []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("trailing JSON input")
	}
	return value, nil
}

func Marshal(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
