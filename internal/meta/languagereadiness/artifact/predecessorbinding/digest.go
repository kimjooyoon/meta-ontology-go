package predecessorbinding

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func registryDigest() string {
	return digestJSON(struct {
		Schema      string       `json:"schema"`
		Coordinates []Coordinate `json:"coordinates"`
	}{Schema: RegistrySchema, Coordinates: coordinates})
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 20 && value == hex.EncodeToString(decoded)
}
