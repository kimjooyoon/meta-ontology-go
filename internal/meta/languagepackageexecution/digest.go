package languagepackageexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func seal(report *Report) {
	report.Digest = ""
	report.Digest = digestValue(*report)
}

func Marshal(report Report) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
