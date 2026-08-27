package nonmonotonicrefutationoracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func transitionDigest(value Transition) string {
	value.TransitionDigest = ""
	return digestJSON(value)
}

func reportDigest(value Report) string {
	value.ReportDigest = ""
	return digestJSON(value)
}
