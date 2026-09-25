package languagediagnosticprovenancebinding

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func artifactDigest(artifact Artifact) string {
	artifact.ArtifactDigest = ""
	return digestValue(artifact)
}
