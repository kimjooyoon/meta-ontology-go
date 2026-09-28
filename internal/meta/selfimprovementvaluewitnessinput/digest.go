package selfimprovementvaluewitnessinput

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func executionInputDigest(value ExecutionInput) string {
	value.Digest = ""
	// CandidateDigest is a mirror binding populated after the candidate digest
	// is known. Excluding it avoids a digest cycle while Validate still checks
	// that the mirror equals the candidate bound by the enclosing artifact.
	value.CandidateDigest = ""
	return digestJSON(value)
}

func Digest(value ExecutionInput) string { return executionInputDigest(value) }

func CorpusDigest(corpus []ValueCase) string { return digestJSON(corpus) }
