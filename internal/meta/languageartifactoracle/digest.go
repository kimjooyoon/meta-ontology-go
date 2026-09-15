package languageartifactoracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}

func artifactDigest(artifact sourceArtifact) string {
	artifact.Digest = ""
	return digestValue(artifact)
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}
