package writeset

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestEntries(entries []Entry) string {
	data, _ := json.Marshal(entries)
	return digestBytes(data)
}

func bindReceipt(receipt Receipt) Receipt {
	receipt.ReceiptDigest = ""
	data, _ := json.Marshal(receipt)
	receipt.ReceiptDigest = digestBytes(data)
	return receipt
}
