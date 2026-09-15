package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
)

func verificationDigest(value contract.Verification) string {
	value.Digest = ""
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}
