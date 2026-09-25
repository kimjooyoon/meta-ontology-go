package languagereadiness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func registryDigest() string {
	return digestJSON(struct {
		Schema      string       `json:"schema"`
		Obligations []Obligation `json:"obligations"`
	}{ContractSchema, obligations})
}

func finalize(snapshot Snapshot) Snapshot {
	snapshot.Digest = ""
	snapshot.Digest = digestJSON(snapshot)
	return snapshot
}
