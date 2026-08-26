package actionability

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Marshal(report Report) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}
