package artifact

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
)

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func snapshotDigest(snapshot readiness.Snapshot) string {
	snapshot.Digest = ""
	return digestJSON(snapshot)
}

func seal(receipt Receipt) Receipt {
	receipt.ArtifactDigest = ""
	receipt.ArtifactDigest = digestJSON(receipt)
	return receipt
}
