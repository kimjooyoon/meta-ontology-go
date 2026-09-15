package authorizationfoundation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestRaw(raw []byte) string {
	value := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(value[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestRaw(raw)
}

func sealReceipt(value Receipt) Receipt {
	value.ReceiptDigest = ""
	value.ReceiptDigest = digestValue(value)
	return value
}

func sealSuite(value Suite) Suite {
	value.SuiteDigest = ""
	value.SuiteDigest = digestValue(value)
	return value
}
