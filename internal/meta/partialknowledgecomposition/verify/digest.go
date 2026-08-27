package verify

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
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func transitionDigest(value claimTransition) string {
	value.Digest = ""
	return digestValue(value)
}

func receiptDigest(value receipt) string {
	value.Digest = ""
	return digestValue(value)
}

func reportDigest(value Report) string {
	value.Digest = ""
	return digestValue(value)
}
