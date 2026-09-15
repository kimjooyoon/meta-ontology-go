package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

type decodedReceipt struct {
	Receipt archivedReceipt
	Payload []byte
}

func (value decodedReceipt) payloadDigest() string {
	if len(value.Payload) == 0 {
		return ""
	}
	sum := sha256.Sum256(value.Payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (value decodedReceipt) payloadBase64() string {
	if len(value.Payload) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(value.Payload)
}
