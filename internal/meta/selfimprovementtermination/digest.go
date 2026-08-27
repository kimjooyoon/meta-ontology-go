package selfimprovementtermination

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func digestJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func seal(receipt Receipt) Receipt {
	receipt.ReceiptDigest = ""
	receipt.ReplayDigest = ""
	receipt.ReceiptDigest = digestJSON(receipt)
	receipt.ReplayDigest = digestJSON(struct {
		InputDigest, TraceDigest, ReceiptDigest string
	}{receipt.InputDigest, receipt.TraceDigest, receipt.ReceiptDigest})
	return receipt
}

func encodeJSON(value any) ([]byte, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("termination input: "+format, args...)
}
