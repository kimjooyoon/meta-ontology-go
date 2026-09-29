package languagecomparison

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
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(data)
}

func seal(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = digestValue(receipt)
	return receipt
}

func defaultNonClaims() []string {
	return []string{"general language performance ranking", "native compilation cost",
		"production workload performance", "business correctness", "cross-runner improvement"}
}
