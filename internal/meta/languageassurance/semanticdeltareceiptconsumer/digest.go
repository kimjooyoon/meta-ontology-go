package semanticdeltareceiptconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sealReceipt(receipt *Receipt) {
	copy := *receipt
	copy.ReceiptDigest = ""
	receipt.ReceiptDigest = digestValue(copy)
}

func receiptDigestValid(receipt Receipt) bool {
	digest := receipt.ReceiptDigest
	copy := receipt
	copy.ReceiptDigest = ""
	return digest != "" && digest == digestValue(copy)
}
