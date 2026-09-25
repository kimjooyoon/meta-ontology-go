package languagegointeroperationbinding

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func artifactDigest(artifact Artifact) string {
	artifact.ArtifactDigest = ""
	return digestValue(artifact)
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
