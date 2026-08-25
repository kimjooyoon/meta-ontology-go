package languagedebug

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func seal(receipt Receipt) Receipt {
	receipt.Digest = ""
	data, _ := json.Marshal(receipt)
	sum := sha256.Sum256(data)
	receipt.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return receipt
}

func validDigest(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
