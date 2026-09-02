package artifactemit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func ValidSHA256(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func ValidDigest(artifact Artifact) bool {
	observed := artifact.Digest
	if !ValidSHA256(observed) {
		return false
	}
	artifact.Digest = ""
	payload, err := json.Marshal(artifact)
	if err != nil {
		return false
	}
	sum := sha256.Sum256(payload)
	return observed == "sha256:"+hex.EncodeToString(sum[:])
}
