package selfimprovementattestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func seal(receipt *ResolutionReceipt) error {
	receipt.Digest = ""
	digest, err := digestValue(receipt)
	if err != nil {
		return err
	}
	receipt.Digest = digest
	return nil
}
