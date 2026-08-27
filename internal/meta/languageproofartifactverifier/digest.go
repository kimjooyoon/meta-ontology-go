package languageproofartifactverifier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

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

func artifactDigest(artifact Artifact) string {
	artifact.Digest = ""
	return digestValue(artifact)
}

func claimStatementDigest(claim ClaimStatement) string {
	claim.Digest = ""
	return digestValue(claim)
}

func claimStateDigest(claim ClaimResult) string {
	claim.StateDigest = ""
	return digestValue(claim)
}

func transitionDigest(transition ClaimTransition) string {
	transition.Digest = ""
	return digestValue(transition)
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }
func validHead(value string) bool   { return headPattern.MatchString(value) }
