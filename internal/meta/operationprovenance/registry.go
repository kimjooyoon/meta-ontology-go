package operationprovenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ReceiptSchema = "gooo/meta-operation-provenance-receipt/v3"
	ReportSchema  = "gooo/meta-operation-provenance-verification/v3"
	Toolchain     = "go1.27.0"
)

func digestBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal digest subject: %w", err)
	}
	return digestBytes(payload), nil
}

func sealReceipt(receipt Receipt) (Receipt, error) {
	receipt.Digest = ""
	digest, err := digestValue(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.Digest = digest
	return receipt, nil
}

// SealReceipt is used by the CI harness after attaching its isolated
// before/after observation to a producer receipt.
func SealReceipt(receipt Receipt) (Receipt, error) { return sealReceipt(receipt) }

func digestText(value string) string { return digestBytes([]byte(value)) }
