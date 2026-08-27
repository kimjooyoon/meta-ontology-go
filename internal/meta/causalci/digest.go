package causalci

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func transitionDigest(value TransitionEvidence) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func receiptDigest(value Receipt) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
