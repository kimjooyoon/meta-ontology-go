package selfimprovementloop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
