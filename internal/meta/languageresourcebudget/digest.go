package languageresourcebudget

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
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(payload)
}

func sealTransition(value ClaimTransition) ClaimTransition {
	value.Digest = ""
	value.Digest = digestValue(value)
	return value
}

func sealReport(value Report) Report {
	value.Digest = ""
	value.Digest = digestValue(value)
	return value
}
