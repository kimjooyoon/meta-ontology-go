package languageproofartifact

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

func evidenceDigest(evidence Evidence) string {
	evidence.EvidenceDigest = ""
	return digestValue(evidence)
}

func claimStatementDigest(claim ClaimStatement) string {
	claim.Digest = ""
	return digestValue(claim)
}

func artifactDigest(artifact Artifact) string {
	artifact.Digest = ""
	return digestValue(artifact)
}
